// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.20;

/// @title C2PAManifestAnchorWithCertRegistry
/// @notice Anchors C2PA manifest hashes on-chain and restricts anchoring
///         to a registry of trusted X.509 certificate fingerprints.
///
/// @dev
///  - This contract does NOT parse or validate X.509 certificates.
///    That validation (chain, revocation, validity period, etc.) must
///    be performed off-chain according to the C2PA specification.
///  - Off-chain, you MUST:
///      * Produce a C2PA-compliant manifest.
///      * Sign it with a key backed by an X.509 certificate.
///      * Validate that certificate chain.
///      * Compute `certFingerprint = sha256(derEncodedCertificate)`.
///      * Compute `manifestHash = sha256(manifestBytes)`.
///  - On-chain, this contract:
///      * Maintains a registry mapping `certFingerprint -> attestor address`.
///      * Requires that only registered, active certificates may anchor.
///      * Requires an EIP-191 signature from the attestor over `manifestHash`.
contract C2PAManifestAnchorWithCertRegistry {
    // ============================================================
    // Ownership / Access Control (simple Ownable pattern)
    // ============================================================

    address public owner;

    event OwnershipTransferred(address indexed previousOwner, address indexed newOwner);

    modifier onlyOwner() {
        _onlyOwner();
        _;
    }

    function _onlyOwner() internal view {
        require(msg.sender == owner, "C2PA: caller is not the owner");
    }

    constructor() {
        owner = msg.sender;
        emit OwnershipTransferred(address(0), msg.sender);
    }

    /// @notice Transfer ownership of the contract to a new address.
    function transferOwnership(address newOwner) external onlyOwner {
        require(newOwner != address(0), "C2PA: new owner is zero address");
        emit OwnershipTransferred(owner, newOwner);
        owner = newOwner;
    }

    // ============================================================
    // Certificate Registry
    // ============================================================

    /// @notice Representation of a registered X.509 certificate (by fingerprint).
    /// @dev
    ///  - `fingerprint` SHOULD be sha256 over the DER-encoded certificate.
    ///  - `attestor` is the EVM address that must sign manifest hashes (EIP-191).
    struct Certificate {
        bytes32 fingerprint; // sha256(derEncodedCertificate)
        address attestor;    // EVM address associated with this certificate/key
        bool active;         // whether this certificate is currently allowed to anchor
        string issuer;       // optional, human-readable issuer (for convenience)
        string subject;      // optional, human-readable subject (for convenience)
    }

    /// @dev Mapping from certificate fingerprint to certificate metadata.
    mapping(bytes32 => Certificate) private certificates;

    /// @dev Optional reverse index: attestor address => list of cert fingerprints.
    mapping(address => bytes32[]) private certsByAttestor;

    event CertificateRegistered(bytes32 indexed fingerprint, address indexed attestor);
    event CertificateStatusChanged(bytes32 indexed fingerprint, bool active);

    /// @notice Register or update a certificate fingerprint and bind it to an attestor address.
    /// @param fingerprint SHA-256 (or agreed hash) of the DER-encoded certificate.
    /// @param attestor EVM address that will sign C2PA manifest hashes for this certificate.
    /// @param issuer Human-readable issuer string (optional; off-chain convenience only).
    /// @param subject Human-readable subject string (optional; off-chain convenience only).
    function registerCertificate(
        bytes32 fingerprint,
        address attestor,
        string calldata issuer,
        string calldata subject
    ) external onlyOwner {
        require(fingerprint != bytes32(0), "C2PA: fingerprint is zero");
        require(attestor != address(0), "C2PA: attestor is zero address");

        Certificate storage cert = certificates[fingerprint];

        // If it's a new certificate, add it to the per-attestor list
        if (cert.attestor == address(0)) {
            certsByAttestor[attestor].push(fingerprint);
        }

        cert.fingerprint = fingerprint;
        cert.attestor = attestor;
        cert.active = true;
        cert.issuer = issuer;
        cert.subject = subject;

        emit CertificateRegistered(fingerprint, attestor);
        emit CertificateStatusChanged(fingerprint, true);
    }

    /// @notice Activate or deactivate an existing certificate.
    /// @param fingerprint The fingerprint of the certificate to update.
    /// @param active New active status (true = allowed to anchor, false = blocked).
    function setCertificateStatus(bytes32 fingerprint, bool active) external onlyOwner {
        Certificate storage cert = certificates[fingerprint];
        require(cert.attestor != address(0), "C2PA: unknown certificate");

        cert.active = active;
        emit CertificateStatusChanged(fingerprint, active);
    }

    /// @notice Get certificate metadata by fingerprint.
    function getCertificate(bytes32 fingerprint) external view returns (Certificate memory) {
        return certificates[fingerprint];
    }

    /// @notice Get all certificate fingerprints associated with an attestor address.
    function getCertificatesByAttestor(address attestor) external view returns (bytes32[] memory) {
        return certsByAttestor[attestor];
    }

    /// @notice Check if a certificate fingerprint is active and known.
    function isCertificateActive(bytes32 fingerprint) public view returns (bool) {
        Certificate storage cert = certificates[fingerprint];
        return cert.attestor != address(0) && cert.active;
    }

    // ============================================================
    // C2PA Manifest Anchoring
    // ============================================================

    /// @notice On-chain record of a single attestation to a C2PA manifest hash.
    struct ManifestAttestation {
        bytes32 manifestHash;     // sha256(manifestBytes) per C2PA hard-binding
        bytes32 certFingerprint;  // X.509 cert fingerprint used by this attestor
        address attestor;         // EVM address that signed (bound to the cert)
        uint64 timestamp;         // block timestamp when anchored
        string manifestUri;       // URI where the manifest can be fetched (HTTPS/IPFS/etc.)
        string assetId;           // C2PA asset identifier / binding ID (optional)
    }

    /// @dev Mapping from manifest hash to all attestations for that manifest.
    mapping(bytes32 => ManifestAttestation[]) private attestationsByHash;

    event ManifestAnchored(
        bytes32 indexed manifestHash,
        bytes32 indexed certFingerprint,
        address indexed attestor,
        string manifestUri,
        string assetId,
        uint64 timestamp
    );

    /// @notice Anchor a C2PA manifest hash using a registered, active certificate.
    ///
    /// @dev Off-chain flow:
    ///  1. Validate C2PA manifest & X.509 certificate chain.
    ///  2. Compute `manifestHash = sha256(manifestBytes)`.
    ///  3. Compute `certFingerprint = sha256(derEncodedCertificate)`.
    ///  4. EVM attestor address (associated with that cert) signs `manifestHash`
    ///     using EIP-191: "\x19Ethereum Signed Message:\n32" || manifestHash.
    ///
    /// On-chain:
    ///  1. Ensure certificate is known and active.
    ///  2. Verify the EIP-191 signature against the attestor address.
    ///  3. Store the attestation and emit ManifestAnchored.
    ///
    /// @param manifestHash SHA-256 digest of the C2PA manifest bytes.
    /// @param manifestUri URI from which the manifest can be retrieved.
    /// @param assetId C2PA asset identifier or soft-binding identifier (optional, free-form).
    /// @param certFingerprint Fingerprint of the registered certificate being used.
    /// @param signature EIP-191 signature over `manifestHash` from the cert's attestor address.
    function anchorManifest(
        bytes32 manifestHash,
        string calldata manifestUri,
        string calldata assetId,
        bytes32 certFingerprint,
        bytes calldata signature
    ) external returns (bool) {
        require(manifestHash != bytes32(0), "C2PA: manifestHash is zero");

        Certificate memory cert = certificates[certFingerprint];
        require(cert.attestor != address(0), "C2PA: unknown certificate");
        require(cert.active, "C2PA: certificate not active");

        // Verify EIP-191 signature from the cert's attestor over the manifest hash
        require(
            _verifySignature(manifestHash, signature, cert.attestor),
            "C2PA: invalid attestor signature"
        );

        ManifestAttestation memory att = ManifestAttestation({
            manifestHash: manifestHash,
            certFingerprint: certFingerprint,
            attestor: cert.attestor,
            timestamp: uint64(block.timestamp),
            manifestUri: manifestUri,
            assetId: assetId
        });

        attestationsByHash[manifestHash].push(att);

        emit ManifestAnchored(
            manifestHash,
            certFingerprint,
            cert.attestor,
            manifestUri,
            assetId,
            att.timestamp
        );

        return true;
    }

    /// @notice Get all attestations for a given manifest hash.
    function getManifestAttestations(bytes32 manifestHash)
        external
        view
        returns (ManifestAttestation[] memory)
    {
        return attestationsByHash[manifestHash];
    }

    /// @notice Get how many attestations exist for a given manifest hash.
    function getManifestAttestationCount(bytes32 manifestHash) external view returns (uint256) {
        return attestationsByHash[manifestHash].length;
    }

    // ============================================================
    // Signature Helpers (EIP-191 over manifestHash)
    // ============================================================

    /// @notice Verify an EIP-191 signature over a manifest hash.
    /// @param manifestHash SHA-256 digest of the manifest bytes.
    /// @param signature EIP-191 signature: 65 bytes (r,s,v).
    /// @param expectedSigner EVM address expected to have signed.
    function _verifySignature(
        bytes32 manifestHash,
        bytes memory signature,
        address expectedSigner
    ) internal pure returns (bool) {
        // EIP-191: "\x19Ethereum Signed Message:\n32" || manifestHash
        bytes32 messageHash;
        // solhint-disable-next-line no-inline-assembly
        assembly {
            // Create the EIP-191 message prefix
            let prefix := "\x19Ethereum Signed Message:\n32"
            // Allocate memory for the concatenated message
            let ptr := mload(0x40)
            // Copy prefix (20 bytes)
            mstore(ptr, prefix)
            // Copy manifestHash (32 bytes) after prefix
            mstore(add(ptr, 20), manifestHash)
            // Calculate keccak256 over the 52-byte message (20 + 32)
            messageHash := keccak256(ptr, 52)
        }

        address recoveredSigner = _recoverSigner(messageHash, signature);
        return recoveredSigner == expectedSigner;
    }

    /// @notice Recover signer from an ECDSA signature over a given message hash.
    /// @param messageHash keccak256-encoded EIP-191 message hash.
    /// @param signature ECDSA signature (65 bytes: r (32) | s (32) | v (1)).
    function _recoverSigner(
        bytes32 messageHash,
        bytes memory signature
    ) internal pure returns (address) {
        require(signature.length == 65, "C2PA: bad signature length");

        bytes32 r;
        bytes32 s;
        uint8 v;

        // solhint-disable-next-line no-inline-assembly
        assembly {
            r := mload(add(signature, 32))
            s := mload(add(signature, 64))
            v := byte(0, mload(add(signature, 96)))
        }

        if (v < 27) {
            v += 27;
        }

        require(v == 27 || v == 28, "C2PA: bad v value");

        return ecrecover(messageHash, v, r, s);
    }
}
