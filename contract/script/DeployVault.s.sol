// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import {Vault} from "../src/Vault.sol";
import {Script} from "forge-std/Script.sol";
import {RexToken} from "../src/RexToken.sol";
import {DeployRexToken} from "./DeployRexToken.s.sol";

contract DeployVault is Script {
    function run() external returns (Vault) {
        DeployRexToken deployRexToken = new DeployRexToken();
        RexToken token = deployRexToken.run();
        vm.startBroadcast();

        Vault vault = new Vault(address(token));
        vm.stopBroadcast();
        return vault;
    }
}
