// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import {NFT} from "../src/NFT.sol";
import {Script} from "forge-std/Script.sol";

contract BlueArchive is NFT {
    constructor()
        NFT(
            "BlueArchive",
            "BA",
            "ipfs://QmRdV3K8ypVqquSBQftT6eSvoN6cW7tZRjhePMnKt777V7"
        )
    {}
}

contract DeployNFT is Script {
    function run() external returns (NFT) {
        vm.startBroadcast();
        BlueArchive nft = new BlueArchive();
        vm.stopBroadcast();

        // // 保存合约 ABI
        // string memory abiFile = "web/src/contracts/NFT.json";
        // string memory ABI = vm.readFile("out/NFT.sol/NFT.json");
        // vm.writeFile(abiFile, ABI);
        return nft;
    }
}
