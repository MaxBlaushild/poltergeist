// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.20;

/// @title C2PAManifestAnchor
/// @notice Anchors C2PA manifest hashes on-chain after off-chain validation.
/// @dev
///  - This contract does NOT parse or validate X.509 certificates or C2PA manifests.
///  - Your backend is responsible for:
///      * Validating the C2PA manifest and X.509 certificate chain.
///      * Computing `manifestHash = sha256(manifestBytes)`.
///      * Computing `certFingerprint = sha256(derEncodedCertificate)`.
///      * Ensuring only approved certificates are used.
///  - This contract:
///      * Maintains a registry of known certificate fingerprints.
///      * Only allows the `owner` (your backend wallet / multisig) to:
///          - register / update certificates
///          - anchor manifest hashes
///      * Stores immutable records of (manifestHash, certFingerprint, uri, assetId, timestamp).
contract C2PAManifestAnchor {
    // ============================================================
    // Ownership / Access Control (simple Ownable pattern)
    // ============================================================

    address public owner;

    event OwnershipTransferred(address indexed previousOwner, address indexed newOwner);

    modifier onlyOwner() {
        _onlyOwner();
        _;
    }

    function _onlyOwner() internal {
        require(msg.sender == owner, "C2PA: caller is not the owner");
    }

    constructor() {
        owner = msg.sender;
        emit OwnershipTransferred(address(0), msg.sender);
    }

    /// @notice Transfer contract ownership (e.g., to a multisig or new backend).
    function transferOwnership(address newOwner) external onlyOwner {
        require(newOwner != address(0), "C2PA: new owner is zero address");
        emit OwnershipTransferred(owner, newOwner);
        owner = newOwner;
    }

    // ============================================================
    // Certificate Registry (by fingerprint)
    // ============================================================

    /// @notice Representation of a registered X.509 certificate (by fingerprint).
    /// @dev
    ///  - `fingerprint` SHOULD be sha256(derEncodedCertificate).
    struct Certificate {
        bytes32 fingerprint; // sha256(derEncodedCertificate)
        bool active;         // whether this certificate is currently allowed
        string issuer;       // optional, human-readable issuer (off-chain convenience)
        string subject;      // optional, human-readable subject (off-chain convenience)
    }

    /// @dev Mapping from certificate fingerprint to certificate metadata.
    mapping(bytes32 => Certificate) private certificates;

    /// @dev Optional list of all fingerprints, for off-chain enumeration.
    bytes32[] private allCertificateFingerprints;

    event CertificateRegistered(bytes32 indexed fingerprint);
    event CertificateStatusChanged(bytes32 indexed fingerprint, bool active);

    /// @notice Register or update a certificate fingerprint.
    /// @param fingerprint SHA-256 (or agreed hash) of the DER-encoded certificate.
    /// @param issuer Human-readable issuer string (optional).
    /// @param subject Human-readable subject string (optional).
    function registerCertificate(
        bytes32 fingerprint,
        string calldata issuer,
        string calldata subject
    ) external onlyOwner {
        require(fingerprint != bytes32(0), "C2PA: fingerprint is zero");

        Certificate storage cert = certificates[fingerprint];

        // If it's a new certificate, track it in the index.
        if (cert.fingerprint == bytes32(0)) {
            allCertificateFingerprints.push(fingerprint);
        }

        cert.fingerprint = fingerprint;
        cert.active = true;
        cert.issuer = issuer;
        cert.subject = subject;

        emit CertificateRegistered(fingerprint);
        emit CertificateStatusChanged(fingerprint, true);
    }

    /// @notice Activate or deactivate an existing certificate.
    /// @param fingerprint Fingerprint of the certificate to update.
    /// @param active New active status (true = allowed, false = blocked).
    function setCertificateStatus(bytes32 fingerprint, bool active) external onlyOwner {
        Certificate storage cert = certificates[fingerprint];
        require(cert.fingerprint != bytes32(0), "C2PA: unknown certificate");

        cert.active = active;
        emit CertificateStatusChanged(fingerprint, active);
    }

    /// @notice Get certificate metadata by fingerprint.
    function getCertificate(bytes32 fingerprint) external view returns (Certificate memory) {
        return certificates[fingerprint];
    }

    /// @notice Return all known certificate fingerprints (for off-chain indexing).
    function getAllCertificateFingerprints() external view returns (bytes32[] memory) {
        return allCertificateFingerprints;
    }

    /// @notice Check if a certificate fingerprint is active and known.
    function isCertificateActive(bytes32 fingerprint) public view returns (bool) {
        Certificate storage cert = certificates[fingerprint];
        return cert.fingerprint != bytes32(0) && cert.active;
    }

    // ============================================================
    // C2PA Manifest Anchoring
    // ============================================================

    /// @notice On-chain record of a single attestation to a C2PA manifest hash.
    struct ManifestAttestation {
        bytes32 manifestHash;    // sha256(manifestBytes) per C2PA hard binding
        bytes32 certFingerprint; // fingerprint of the X.509 certificate used (off-chain)
        uint64 timestamp;        // block timestamp when anchored
        string manifestUri;      // where the manifest can be retrieved (HTTPS/IPFS/etc.)
        string assetId;          // C2PA asset identifier / binding ID (optional)
    }

    /// @dev Mapping from manifest hash to all attestations for that manifest.
    mapping(bytes32 => ManifestAttestation[]) private attestationsByHash;

    event ManifestAnchored(
        bytes32 indexed manifestHash,
        bytes32 indexed certFingerprint,
        uint64 timestamp,
        string manifestUri,
        string assetId
    );

    /// @notice Anchor a C2PA manifest hash using an active, registered certificate.
    ///
    /// @dev Off-chain responsibilities BEFORE calling this:
    ///  1. Validate C2PA manifest + X.509 chain according to the C2PA spec.
    ///  2. Compute `manifestHash = sha256(manifestBytes)`.
    ///  3. Compute `certFingerprint = sha256(derEncodedCertificate)`.
    ///  4. Confirm that this certificate is allowed by your policy.
    ///
    /// @param manifestHash SHA-256 digest of the C2PA manifest bytes.
    /// @param manifestUri URI where the manifest can be retrieved.
    /// @param assetId C2PA asset identifier or soft-binding identifier (optional).
    /// @param certFingerprint Fingerprint of the registered certificate used to sign the manifest.
    function anchorManifest(
        bytes32 manifestHash,
        string calldata manifestUri,
        string calldata assetId,
        bytes32 certFingerprint
    ) external onlyOwner returns (bool) {
        require(manifestHash != bytes32(0), "C2PA: manifestHash is zero");
        require(isCertificateActive(certFingerprint), "C2PA: certificate not active");

        ManifestAttestation memory att = ManifestAttestation({
            manifestHash: manifestHash,
            certFingerprint: certFingerprint,
            timestamp: uint64(block.timestamp),
            manifestUri: manifestUri,
            assetId: assetId
        });

        attestationsByHash[manifestHash].push(att);

        emit ManifestAnchored(
            manifestHash,
            certFingerprint,
            att.timestamp,
            manifestUri,
            assetId
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
}
