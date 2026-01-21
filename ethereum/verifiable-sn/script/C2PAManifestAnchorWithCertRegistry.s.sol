// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.20;

import {Script} from "forge-std/Script.sol";
import {console} from "forge-std/console.sol";
import {C2PAManifestAnchorWithCertRegistry} from "../src/C2PAManifestAnchorWithCertRegistry.sol";

contract C2PManifestAnchorWithCertRegistryVerifierScript is Script {
    function setUp() public {}
    
    function run() public {
        uint256 deployerPrivateKey = vm.envUint("PRIVATE_KEY");
        vm.startBroadcast(deployerPrivateKey);
        
        C2PAManifestAnchorWithCertRegistry c2pManifestAnchorWithCertRegistryVerifier = new C2PAManifestAnchorWithCertRegistry();
        
        console.log("C2PManifestAnchorWithCertRegistryVerifier deployed at:", address(c2pManifestAnchorWithCertRegistryVerifier));
        
        vm.stopBroadcast();
    }
}


