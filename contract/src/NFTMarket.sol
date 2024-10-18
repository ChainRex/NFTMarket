// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import {IERC721} from "@openzeppelin/contracts/token/ERC721/IERC721.sol";
import {IERC20} from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import {IERC20Permit} from "@openzeppelin/contracts/token/ERC20/extensions/IERC20Permit.sol";
import {NFT} from "./NFT.sol";
contract NFTMarket {
    // 交易对
    struct Order {
        address nft;
        uint256 tokenId;
        address token;
        uint256 price;
        address seller;
        uint256 status; // 0: 未售出, 1: 已售出, 2: 已取消
    }

    // 交易对列表
    Order[] public orders;

    // 创建交易对
    function createOrder(
        address nft,
        uint256 tokenId,
        address token,
        uint256 price
    ) public {
        require(price > 0, "Invalid price");
        // 检查是否approve
        require(
            IERC721(nft).getApproved(tokenId) == address(this) ||
                IERC721(nft).isApprovedForAll(msg.sender, address(this)),
            "Not approved"
        );
        orders.push(Order(nft, tokenId, token, price, msg.sender, 0));
    }

    // 取消交易对
    function cancelOrder(uint256 index) public {
        orders[index].status = 2;
    }

    // 购买NFT
    function buyNFT(
        uint256 index,
        uint256 deadline,
        uint8 v,
        bytes32 r,
        bytes32 s
    ) public {
        require(index < orders.length, "Order does not exist");
        Order storage order = orders[index];
        require(order.status == 0, "Order is not available for sale");

        // 尝试执行permit
        try
            IERC20Permit(order.token).permit(
                msg.sender,
                address(this),
                order.price,
                deadline,
                v,
                r,
                s
            )
        {} catch {
            revert("Permit failed");
        }

        // 尝试转账代币
        require(
            IERC20(order.token).transferFrom(
                msg.sender,
                order.seller,
                order.price
            ),
            "Token transfer failed"
        );

        // 尝试转移NFT
        try
            IERC721(order.nft).transferFrom(
                order.seller,
                msg.sender,
                order.tokenId
            )
        {
            order.status = 1;
        } catch {
            revert("NFT transfer failed");
        }
    }

    function getOrderCount() public view returns (uint256) {
        return orders.length;
    }

    function getOrders() public view returns (Order[] memory) {
        return orders;
    }

    // 新增: 部署NFT合约的事件
    event NFTContractDeployed(
        address indexed nftAddress,
        string name,
        string symbol
    );

    // 新增: 部署NFT合约的函数
    function deployNFTContract(
        string memory name,
        string memory symbol,
        string memory tokenIconURI
    ) public returns (address) {
        // 创建新的NFT合约
        NFT newNFTContract = new NFT(name, symbol, tokenIconURI);

        // 发出事件
        emit NFTContractDeployed(address(newNFTContract), name, symbol);

        return address(newNFTContract);
    }
}
