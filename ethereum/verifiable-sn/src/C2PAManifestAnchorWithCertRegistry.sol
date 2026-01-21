// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.13;

/**
 * @title HashVerifier
 * @notice Contract that verifies EIP-191 signatures and stores verified hashes per address
 * @dev Uses ecrecover to verify that a hash was signed by a specific address
 */
contract HashVerifier {
    // Mapping from address to hash to verification status
    mapping(address => mapping(bytes32 => bool)) public verifiedHashes;
    
    // Mapping from address to array of all verified hashes (for enumeration)
    mapping(address => bytes32[]) public addressHashes;
    
    // Event emitted when a hash is successfully verified and stored
    event HashVerified(address indexed signer, bytes32 indexed hash, uint256 timestamp);
    
    /**
     * @notice Verify and store a hash that was signed by a specific address
     * @param hash The hash that was signed
     * @param signature The EIP-191 signature of the hash
     * @param signer The address that should have signed the hash
     * @return success True if the hash was successfully verified and stored
     */
    function submitHash(
        bytes32 hash,
        bytes memory signature,
        address signer
    ) public returns (bool success) {
        // Verify the signature
        require(verifySignature(hash, signature, signer), "HashVerifier: Invalid signature");
        
        // Check if hash is already verified for this address
        require(!verifiedHashes[signer][hash], "HashVerifier: Hash already verified");
        
        // Store the verified hash
        verifiedHashes[signer][hash] = true;
        addressHashes[signer].push(hash);
        
        // Emit event
        emit HashVerified(signer, hash, block.timestamp);
        
        return true;
    }
    
    /**
     * @notice Verify if a hash is verified for a specific address
     * @param signer The address to check
     * @param hash The hash to check
     * @return isVerified True if the hash is verified for the address
     */
    function isHashVerified(address signer, bytes32 hash) public view returns (bool isVerified) {
        return verifiedHashes[signer][hash];
    }
    
    /**
     * @notice Get all verified hashes for an address
     * @param signer The address to query
     * @return hashes Array of all verified hashes for the address
     */
    function getAddressHashes(address signer) public view returns (bytes32[] memory hashes) {
        return addressHashes[signer];
    }
    
    /**
     * @notice Get the count of verified hashes for an address
     * @param signer The address to query
     * @return count Number of verified hashes
     */
    function getHashCount(address signer) public view returns (uint256 count) {
        return addressHashes[signer].length;
    }
    
    /**
     * @notice Verify an EIP-191 signature
     * @param hash The hash that was signed
     * @param signature The signature (65 bytes: r, s, v)
     * @param signer The expected signer address
     * @return isValid True if the signature is valid and matches the signer
     */
    function verifySignature(
        bytes32 hash,
        bytes memory signature,
        address signer
    ) public pure returns (bool isValid) {
        // EIP-191: Add Ethereum signed message prefix
        bytes32 messageHash = keccak256(
            abi.encodePacked("\x19Ethereum Signed Message:\n32", hash)
        );
        
        // Recover the signer from the signature
        address recoveredSigner = recoverSigner(messageHash, signature);
        
        // Verify the recovered signer matches the expected signer
        return recoveredSigner == signer;
    }
    
    /**
     * @notice Recover the signer address from a signature
     * @param messageHash The message hash that was signed
     * @param signature The signature (65 bytes: r, s, v)
     * @return signer The recovered signer address
     */
    function recoverSigner(
        bytes32 messageHash,
        bytes memory signature
    ) public pure returns (address signer) {
        // Check signature length (should be 65 bytes: 32 bytes r + 32 bytes s + 1 byte v)
        require(signature.length == 65, "HashVerifier: Invalid signature length");
        
        bytes32 r;
        bytes32 s;
        uint8 v;
        
        // Extract r, s, v from signature
        assembly {
            r := mload(add(signature, 32))
            s := mload(add(signature, 64))
            v := byte(0, mload(add(signature, 96)))
        }
        
        // Handle v being 0 or 1 (add 27 to get 27 or 28)
        if (v < 27) {
            v += 27;
        }
        
        // Verify v is 27 or 28
        require(v == 27 || v == 28, "HashVerifier: Invalid signature v value");
        
        // Recover the signer using ecrecover
        return ecrecover(messageHash, v, r, s);
    }
}

