// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.13;

import {Test} from "forge-std/Test.sol";
import {HashVerifier} from "../src/HashVerifier.sol";

contract HashVerifierTest is Test {
    HashVerifier public hashVerifier;
    
    address public alice;
    address public bob;
    uint256 public alicePrivateKey;
    uint256 public bobPrivateKey;
    
    function setUp() public {
        hashVerifier = new HashVerifier();
        
        // Create test accounts
        alicePrivateKey = 0xa11ce;
        bobPrivateKey = 0xb0b;
        alice = vm.addr(alicePrivateKey);
        bob = vm.addr(bobPrivateKey);
    }
    
    function test_SubmitHash_Success() public {
        bytes32 hash = keccak256("test hash");
        
        // vm.sign automatically adds EIP-191 prefix, so we pass the raw hash
        (uint8 v, bytes32 r, bytes32 s) = vm.sign(alicePrivateKey, hash);
        bytes memory signature = abi.encodePacked(r, s, v);
        
        // Submit the hash
        bool success = hashVerifier.submitHash(hash, signature, alice);
        assertTrue(success);
        
        // Verify the hash is stored
        assertTrue(hashVerifier.isHashVerified(alice, hash));
        assertEq(hashVerifier.getHashCount(alice), 1);
    }
    
    function test_SubmitHash_InvalidSignature() public {
        bytes32 hash = keccak256("test hash");
        
        // Sign with Alice's key but claim Bob signed it
        (uint8 v, bytes32 r, bytes32 s) = vm.sign(alicePrivateKey, hash);
        bytes memory signature = abi.encodePacked(r, s, v);
        
        // Should fail because signature doesn't match Bob
        vm.expectRevert("HashVerifier: Invalid signature");
        hashVerifier.submitHash(hash, signature, bob);
    }
    
    function test_SubmitHash_WrongSigner() public {
        bytes32 hash = keccak256("test hash");
        
        // Sign with Alice's key
        (uint8 v, bytes32 r, bytes32 s) = vm.sign(alicePrivateKey, hash);
        bytes memory signature = abi.encodePacked(r, s, v);
        
        // Try to submit with wrong signer address
        vm.expectRevert("HashVerifier: Invalid signature");
        hashVerifier.submitHash(hash, signature, bob);
    }
    
    function test_SubmitHash_Duplicate() public {
        bytes32 hash = keccak256("test hash");
        
        // Sign the hash with Alice's private key
        (uint8 v, bytes32 r, bytes32 s) = vm.sign(alicePrivateKey, hash);
        bytes memory signature = abi.encodePacked(r, s, v);
        
        // Submit the hash first time
        hashVerifier.submitHash(hash, signature, alice);
        
        // Try to submit again - should fail
        vm.expectRevert("HashVerifier: Hash already verified");
        hashVerifier.submitHash(hash, signature, alice);
    }
    
    function test_MultipleHashes() public {
        bytes32 hash1 = keccak256("hash 1");
        bytes32 hash2 = keccak256("hash 2");
        bytes32 hash3 = keccak256("hash 3");
        
        // Sign and submit multiple hashes
        (uint8 v1, bytes32 r1, bytes32 s1) = vm.sign(alicePrivateKey, hash1);
        bytes memory sig1 = abi.encodePacked(r1, s1, v1);
        hashVerifier.submitHash(hash1, sig1, alice);
        
        (uint8 v2, bytes32 r2, bytes32 s2) = vm.sign(alicePrivateKey, hash2);
        bytes memory sig2 = abi.encodePacked(r2, s2, v2);
        hashVerifier.submitHash(hash2, sig2, alice);
        
        (uint8 v3, bytes32 r3, bytes32 s3) = vm.sign(alicePrivateKey, hash3);
        bytes memory sig3 = abi.encodePacked(r3, s3, v3);
        hashVerifier.submitHash(hash3, sig3, alice);
        
        // Verify all hashes are stored
        assertTrue(hashVerifier.isHashVerified(alice, hash1));
        assertTrue(hashVerifier.isHashVerified(alice, hash2));
        assertTrue(hashVerifier.isHashVerified(alice, hash3));
        assertEq(hashVerifier.getHashCount(alice), 3);
        
        // Verify Bob's hashes are separate
        assertFalse(hashVerifier.isHashVerified(bob, hash1));
        assertEq(hashVerifier.getHashCount(bob), 0);
    }
    
    function test_GetAddressHashes() public {
        bytes32 hash1 = keccak256("hash 1");
        bytes32 hash2 = keccak256("hash 2");
        
        // Sign and submit hashes
        (uint8 v1, bytes32 r1, bytes32 s1) = vm.sign(alicePrivateKey, hash1);
        bytes memory sig1 = abi.encodePacked(r1, s1, v1);
        hashVerifier.submitHash(hash1, sig1, alice);
        
        (uint8 v2, bytes32 r2, bytes32 s2) = vm.sign(alicePrivateKey, hash2);
        bytes memory sig2 = abi.encodePacked(r2, s2, v2);
        hashVerifier.submitHash(hash2, sig2, alice);
        
        // Get all hashes for Alice
        bytes32[] memory hashes = hashVerifier.getAddressHashes(alice);
        assertEq(hashes.length, 2);
        assertEq(hashes[0], hash1);
        assertEq(hashes[1], hash2);
    }
    
    function test_EventEmission() public {
        bytes32 hash = keccak256("test hash");
        
        // Sign the hash
        (uint8 v, bytes32 r, bytes32 s) = vm.sign(alicePrivateKey, hash);
        bytes memory signature = abi.encodePacked(r, s, v);
        
        // Expect event emission
        vm.expectEmit(true, true, false, true);
        emit HashVerifier.HashVerified(alice, hash, block.timestamp);
        
        // Submit the hash
        hashVerifier.submitHash(hash, signature, alice);
    }
    
    function test_VerifySignature() public {
        bytes32 hash = keccak256("test hash");
        
        // Sign the hash
        (uint8 v, bytes32 r, bytes32 s) = vm.sign(alicePrivateKey, hash);
        bytes memory signature = abi.encodePacked(r, s, v);
        
        // Verify signature function
        bool isValid = hashVerifier.verifySignature(hash, signature, alice);
        assertTrue(isValid);
        
        // Verify wrong signer fails
        bool isInvalid = hashVerifier.verifySignature(hash, signature, bob);
        assertFalse(isInvalid);
    }
}

