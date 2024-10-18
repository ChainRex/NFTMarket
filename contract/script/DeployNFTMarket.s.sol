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
            memory addressFile = "../web/src/contracts/NFTMarket-address.json";
        vm.writeJson(
            string(
                abi.encodePacked(
                    '{"address": "',
                    vm.toString(address(nftMarket)),
                    '"}'
                )
            ),
            addressFile
        );

        // 保存合约 ABI
        string memory abiFile = "../web/src/contracts/NFTMarket-abi.json";
        string memory ABI = vm.readFile("out/NFTMarket.sol/NFTMarket.json");
        vm.writeFile(abiFile, ABI);
        return nftMarket;
    }
}
