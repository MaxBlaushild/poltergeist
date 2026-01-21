// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.13;

import {Script} from "forge-std/Script.sol";
import {HashVerifier} from "../src/HashVerifier.sol";

contract HashVerifierScript is Script {
    function setUp() public {}
    
    function run() public {
        uint256 deployerPrivateKey = vm.envUint("PRIVATE_KEY");
        vm.startBroadcast(deployerPrivateKey);
        
        HashVerifier hashVerifier = new HashVerifier();
        
        console.log("HashVerifier deployed at:", address(hashVerifier));
        
        vm.stopBroadcast();
    }
}

