// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import {ERC721URIStorage, ERC721} from "@openzeppelin/contracts/token/ERC721/extensions/ERC721URIStorage.sol";

contract NFT is ERC721URIStorage {
    uint256 private _nextTokenId;
    uint256 public totalSupply;
    string public tokenIconURI;

    constructor(
        string memory name,
        string memory symbol,
        string memory tokenIconURI_
    ) ERC721(name, symbol) {
        tokenIconURI = tokenIconURI_;
    }

    function mint(
        address to,
        string memory tokenURI_
    ) public returns (uint256) {
        uint256 tokenId = _nextTokenId++;
        _safeMint(to, tokenId);
        _setTokenURI(tokenId, tokenURI_);
        totalSupply++;
        return tokenId;
    }

    function getTokenURIList() public view returns (string[] memory) {
        string[] memory tokenURIs = new string[](totalSupply);
        for (uint256 i = 0; i < totalSupply; i++) {
            tokenURIs[i] = tokenURI(i);
        }
        return tokenURIs;
    }
}
