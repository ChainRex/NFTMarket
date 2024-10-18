// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import {RexToken} from "../src/RexToken.sol";
import {Script} from "forge-std/Script.sol";

contract DeployRexToken is Script {
    function run() external returns (RexToken) {
        vm.startBroadcast();
        RexToken token = new RexToken();
        vm.stopBroadcast();
        return token;
    }
}
