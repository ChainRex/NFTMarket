// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.13;

import {Script} from "forge-std/Script.sol";
import {NFTMarket} from "../src/NFTMarket.sol";

contract DeployNFTMarket is Script {
    function run() public returns (NFTMarket) {
        vm.startBroadcast();

        NFTMarket nftMarket = new NFTMarket();

        vm.stopBroadcast();

        // 保存合约地址
        string
            memory webAddressFile = "../web/src/contracts/NFTMarket-address.json";
        string
            memory backendAddressFile = "../backend/contracts/NFTMarket-address.json";
        string memory addressJson = string(
            abi.encodePacked(
                '{"address": "',
                vm.toString(address(nftMarket)),
                '"}'
            )
        );
        vm.writeJson(addressJson, webAddressFile);
        vm.writeJson(addressJson, backendAddressFile);

        // 保存合约 ABI
        string memory webAbiFile = "../web/src/contracts/NFTMarket-abi.json";
        string
            memory backendAbiFile = "../backend/contracts/NFTMarket-abi.json";
        string memory ABI = vm.readFile("out/NFTMarket.sol/NFTMarket.json");
        vm.writeFile(webAbiFile, ABI);
        vm.writeFile(backendAbiFile, ABI);

        return nftMarket;
    }
}
