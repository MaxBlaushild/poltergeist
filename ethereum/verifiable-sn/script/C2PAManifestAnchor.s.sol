// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.20;

import {Script} from "forge-std/Script.sol";
import {console} from "forge-std/console.sol";
import {C2PAManifestAnchor} from "../src/C2PAManifestAnchor.sol";

contract C2PManifestAnchorScript is Script {
    function setUp() public {}
    
    function run() public {
        uint256 deployerPrivateKey = vm.envUint("PRIVATE_KEY");
        vm.startBroadcast(deployerPrivateKey);
        
        C2PAManifestAnchor c2pManifestAnchor = new C2PAManifestAnchor();
        
        console.log("C2PManifestAnchor deployed at:", address(c2pManifestAnchor));
        
        vm.stopBroadcast();
    }
}


