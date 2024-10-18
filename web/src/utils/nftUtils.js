import { ethers } from 'ethers';
import axios from 'axios';
import { getProvider } from './contract';
import store from '../store'; // 导入 store
import NFTABI from '../contracts/NFT.json';

const IPFS_GATEWAYS = [
    'https://gateway.pinata.cloud/ipfs/',
];

export const getNFTImageUrl = async (nftAddress, tokenId) => {
    // 先检查 store
    const storedImageUrl = store.state.nftImageUrls[`${nftAddress}-${tokenId}`];
    if (storedImageUrl) {
        console.log('从 store 获取 NFT 图像 URL:', storedImageUrl);
        return storedImageUrl;
    }

    try {
        const provider = await getProvider();
        const nftContract = new ethers.Contract(
            nftAddress,
            ['function tokenURI(uint256 tokenId) view returns (string)'],
            provider
        );
        const tokenURI = await nftContract.tokenURI(tokenId);

        for (const gateway of IPFS_GATEWAYS) {
            try {
                const ipfsUrl = tokenURI.replace('ipfs://', gateway);
                const response = await axios.get(ipfsUrl, { timeout: 5000 });
                const metadata = response.data;
                const imageUrl = metadata.image.replace('ipfs://', gateway);

                // 更新 store
                store.commit('setNFTImageUrl', { nftAddress, tokenId, imageUrl });

                return imageUrl;
            } catch (error) {
                console.error(`使用网关 ${gateway} 获取失败:`, error);
            }
        }
        throw new Error('所有 IPFS 网关都失败了');
    } catch (error) {
        console.error('获取 NFT 图像 URL 失败:', error);
        return null;
    }
};

export const getTokenInfo = async (tokenAddress) => {
    // 先检查 store
    const storedTokenInfo = store.state.tokenInfo[tokenAddress];
    if (storedTokenInfo) {
        console.log('从 store 获取代币信息:', storedTokenInfo);
        return storedTokenInfo;
    }

    const provider = await getProvider();
    const tokenContract = new ethers.Contract(
        tokenAddress,
        ['function name() view returns (string)', 'function symbol() view returns (string)'],
        provider
    );

    try {
        const [name, symbol] = await Promise.all([
            tokenContract.name(),
            tokenContract.symbol()
        ]);
        const tokenInfo = { name, symbol };

        // 更新 store
        store.commit('setTokenInfo', { address: tokenAddress, info: tokenInfo });

        return tokenInfo;
    } catch (error) {
        console.error('获取代币信息失败:', error);
        return { name: 'Unknown', symbol: 'Unknown' };
    }
};

export const getNFTName = async (nftAddress) => {
    // 先检查 store
    const storedName = store.state.nftNames[nftAddress];
    if (storedName) {
        console.log('从 store 获取 NFT 名称:', storedName);
        return storedName;
    }

    const provider = await getProvider();
    const nftContract = new ethers.Contract(
        nftAddress,
        ['function name() view returns (string)'],
        provider
    );

    try {
        const name = await nftContract.name();
        // 更新 store
        store.commit('setNFTInfo', { address: nftAddress, field: 'name', value: name });
        return name;
    } catch (error) {
        console.error('获取 NFT 名称失败:', error);
        return 'Unknown Collection';
    }
};

export const getNFTTokenIconURI = async (nftAddress) => {
    // 先检查 store
    const storedIconURI = store.state.nftIconURIs[nftAddress];
    if (storedIconURI) {
        console.log('从 store 获取 NFT tokenIconURI:', storedIconURI);
        return storedIconURI;
    }

    const provider = await getProvider();
    const nftContract = new ethers.Contract(
        nftAddress,
        ['function tokenIconURI() view returns (string)'],
        provider
    );

    try {
        const tokenIconURI = await nftContract.tokenIconURI();
        // 更新 store
        store.commit('setNFTInfo', { address: nftAddress, field: 'iconURI', value: tokenIconURI });
        return tokenIconURI;
    } catch (error) {
        console.error('获取 NFT tokenIconURI 失败:', error);
        return null;
    }
};

export const getIPFSUrl = (ipfsUri) => {
    console.log('ipfsUri', ipfsUri);
    if (!ipfsUri) return null;
    for (const gateway of IPFS_GATEWAYS) {
        if (ipfsUri.startsWith('ipfs://')) {
            return ipfsUri.replace('ipfs://', gateway);
        }
    }
    return ipfsUri; // 如果不是 ipfs:// 链接,则返回原始 URI
};

export const getTokenURI = async (nftAddress, tokenId) => {
    // 先检查 store
    const storedTokenURI = store.state.tokenURIs[`${nftAddress}-${tokenId}`];
    if (storedTokenURI) {
        console.log('从 store 获取 tokenURI:', storedTokenURI);
        return storedTokenURI;
    }

    try {
        const provider = await getProvider();
        const nftContract = new ethers.Contract(nftAddress, NFTABI.abi, provider);
        const tokenURI = await nftContract.tokenURI(tokenId);

        // 更新 store
        store.commit('setTokenURI', { nftAddress, tokenId, tokenURI });

        return tokenURI;
    } catch (error) {
        console.error('获取 tokenURI 失败:', error);
        return null;
    }
};
