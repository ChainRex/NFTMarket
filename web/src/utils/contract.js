import { ethers } from 'ethers';
import contractAddress from '../contracts/NFTMarket-address.json';
import contractABI from '../contracts/NFTMarket-abi.json';
import { ElMessage } from 'element-plus';
import { handleGlobalError } from './errorHandler';

const address = contractAddress.address;
const abi = contractABI.abi;

let provider;
let signer;
let contract;

const FALLBACK_RPC_URL = "https://polygon-amoy.g.alchemy.com/v2/oUhC0fClZFJKJ09zzWsqj65EFq3X01y0";
const EXPECTED_CHAIN_ID = 80002; // Polygon Amoy 测试网的 chainId

// 缓存的 provider 实例
let cachedProvider = null;

// 添加这些常量定义
const MAX_RETRIES = 3;
const RETRY_DELAY = 2000; // 2 seconds

const sleep = (ms) => new Promise(resolve => setTimeout(resolve, ms));

export async function getProvider() {
    console.log('Current cached provider:', cachedProvider);
    if (cachedProvider) {
        return cachedProvider;
    }

    if (typeof window !== 'undefined' && typeof window.ethereum !== 'undefined') {
        try {
            const tempProvider = new ethers.providers.Web3Provider(window.ethereum);
            const network = await tempProvider.getNetwork();
            console.log('Detected network:', network);
            if (network.chainId === EXPECTED_CHAIN_ID) {
                console.log('User is on Polygon Amoy testnet');
                cachedProvider = tempProvider;
            } else {
                console.warn('用户不在 Polygon Amoy 测试网，使用只读提供者');
                cachedProvider = new ethers.providers.JsonRpcProvider(FALLBACK_RPC_URL);
            }
        } catch (error) {
            console.error('获取网络信息时出错:', error);
            console.warn('使用只读提供者');
            cachedProvider = new ethers.providers.JsonRpcProvider(FALLBACK_RPC_URL);
        }
    } else {
        console.log('No Web3 detected, using fallback RPC');
        cachedProvider = new ethers.providers.JsonRpcProvider(FALLBACK_RPC_URL);
    }
    console.log('Returning provider:', cachedProvider);
    return cachedProvider;
}

export async function initContract(useWallet = false) {
    try {
        provider = await getProvider();
        if (useWallet && typeof window.ethereum !== 'undefined') {
            await window.ethereum.request({ method: 'eth_requestAccounts' });
            const accounts = await provider.listAccounts();
            if (accounts.length > 0) {
                signer = provider.getSigner();
                contract = new ethers.Contract(address, abi, signer);
            } else {
                console.warn('No accounts found. User might not be connected.');
                contract = new ethers.Contract(address, abi, provider);
            }
        } else {
            contract = new ethers.Contract(address, abi, provider);
        }
        return contract;
    } catch (error) {
        console.error('合约初始化失败:', error);
        handleGlobalError(error);
        throw error;
    }
}

export async function buyNFT(index, deadline, v, r, s) {
    initContract(true)
    if (!contract) await initContract(true);
    if (!signer) throw new Error('需要连接钱包来购买 NFT');
    try {
        const tx = await contract.buyNFT(index, deadline, v, r, s);
        return tx;
    } catch (error) {
        console.error('购买 NFT 时出错:', error);
        handleGlobalError(error);
        throw error;
    }
}

export async function createOrder(nft, tokenId, token, price) {
    initContract(true)
    if (!contract) await initContract(true);
    if (!signer) throw new Error('需要连接钱包来创建订单');
    const tx = await contract.createOrder(nft, tokenId, token, ethers.utils.parseEther(price));
    await tx.wait();
}

export async function cancelOrder(orderId) {
    initContract(true)
    if (!contract) await initContract(true);
    if (!signer) throw new Error('需要连接钱包来取消订单');

    try {
        console.log('正在取消订单...');
        const tx = await contract.cancelOrder(orderId);
        await tx.wait();
        console.log('订单已成功取消');
        return true;
    } catch (error) {
        console.error('取消订单失败:', error);
        handleGlobalError(error);
        throw error;
    }
}

export async function getOrders() {
    if (!contract) await initContract();

    console.log('合约地址:', contract.address);

    try {
        const orders = await contract.getOrders();
        console.log('获取到的订单:', orders);
        return orders;
    } catch (error) {
        console.error('获取单时出错:', error);
        handleGlobalError(error);
        throw error;
    }
}

export async function createOrderWithApprove(nft, tokenId, token, price, signer) {
    if (!contract) {
        contract = await initContract(true);
    }
    if (!signer) {
        throw new Error('需要连接钱包来创建订单');
    }

    const nftContract = new ethers.Contract(nft, [
        'function approve(address to, uint256 tokenId) public',
        'function getApproved(uint256 tokenId) public view returns (address)'
    ], signer);

    let retries = 0;
    while (retries < MAX_RETRIES) {
        try {
            const approvedAddress = await nftContract.getApproved(tokenId);

            if (approvedAddress !== address) {
                console.log('正在授权 NFT...');
                const approveTx = await nftContract.approve(address, tokenId);
                await approveTx.wait();
                console.log('NFT 已授权');
            }

            console.log('正在创建订单...');
            const priceInWei = ethers.utils.parseUnits(price, 18);
            const createOrderTx = await contract.connect(signer).createOrder(nft, tokenId, token, priceInWei);
            await createOrderTx.wait();
            console.log('订单已创建');
            return true;
        } catch (error) {
            console.error(`尝试 ${retries + 1} 失败:`, error);
            if (error.code === 4001 || error.message.includes('user rejected transaction')) {
                ElMessage.info('您取消了交易签名，订单未创建');
                return false; // 用户拒绝交易时返回 false
            }
            if (retries === MAX_RETRIES - 1) {
                throw error; // 如果是最后一次尝试，则抛出错误
            }
            ElMessage.warning(`创建订单失败: ${error.message}。将在 ${RETRY_DELAY / 1000} 秒后重试...`);
            await sleep(RETRY_DELAY); // 等待一段时间后重试
            retries++;
        }
    }
    return false; // 如果所有尝试都失败，返回 false
}

export async function deployNFTContract(name, symbol, tokenIconURI) {
    if (!contract) await initContract(true);
    if (!signer) throw new Error('需要连接钱包来部署 NFT 合约');

    let retries = 0;
    while (retries < MAX_RETRIES) {
        try {
            console.log(`尝试部署 NFT 合约 (尝试 ${retries + 1}/${MAX_RETRIES})`);
            const tx = await contract.deployNFTContract(name, symbol, tokenIconURI);
            console.log('交易已提交，等待确认...');
            const receipt = await tx.wait();
            console.log('交易已确认');
            const deployedEvent = receipt.events.find(e => e.event === 'NFTContractDeployed');
            if (!deployedEvent) {
                throw new Error('未找到 NFTContractDeployed 事件');
            }
            return deployedEvent.args.nftAddress;
        } catch (error) {
            console.error(`部署失败 (尝试 ${retries + 1}/${MAX_RETRIES}):`, error);
            if (error.code === 4001) { // MetaMask 错误码 4001 表示用户拒绝交易
                ElMessage.error('用户拒绝了交易，操作已取消');
                throw error; // 用户拒绝交易，直接抛出错误，不再重试
            }
            if (retries === MAX_RETRIES - 1) {
                throw error; // 如果是最后一次尝试，则抛出错误
            }
            ElMessage.warning(`部署 NFT 合约失败: ${error.message}。将在 ${RETRY_DELAY / 1000} 秒后重试...`);
            await sleep(RETRY_DELAY); // 等待一段时间后重试
            retries++;
        }
    }
}

export function clearProviderCache() {
    cachedProvider = null;
    signer = null;
    contract = null;
}