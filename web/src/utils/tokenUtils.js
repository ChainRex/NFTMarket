import axios from 'axios'
import { ethers } from 'ethers'
import { getNFTImageUrl } from './nftUtils'

const ERC20_ABI = [
    'function balanceOf(address owner) view returns (uint256)',
    'function decimals() view returns (uint8)',
    'function symbol() view returns (string)',
    'function name() view returns (string)'
]

function parseHTML(html) {
    const parser = new DOMParser();
    const doc = parser.parseFromString(html, 'text/html');
    const noEntriesAlert = doc.querySelector('.alert.alert-warning');

    if (noEntriesAlert && noEntriesAlert.textContent.includes('There are no matching entries')) {
        return [];
    }

    const rows = doc.querySelectorAll('div.table-responsive table.table-hover tbody tr');

    const tokenTransfers = [];
    rows.forEach(row => {
        const tds = row.querySelectorAll('td');
        if (tds.length >= 12) {
            const tokenAddress = tds[11].querySelector('a').getAttribute('href').split('?')[0].split('/').pop();
            const tokenName = tds[11].querySelector('.hash-tag').textContent.trim();
            const tokenSymbol = tds[11].querySelector('.text-muted').textContent.replace(/[()]/g, '').trim();
            const amount = tds[10].querySelector('.td_showAmount').textContent.trim();
            const direction = tds[8].textContent.trim();
            const iconSrc = tds[11].querySelector('img').getAttribute('src');

            tokenTransfers.push({
                tokenAddress,
                tokenName,
                tokenSymbol,
                amount,
                direction,
                iconSrc
            });
        }
    });

    return tokenTransfers;
}

async function getAllPages(address) {
    let allTokenTransfers = [];
    let currentPage = 1;
    let hasNextPage = true;

    while (hasNextPage) {
        const url = `/api/tokentxns?a=${address}&ps=100&p=${currentPage}`;
        const response = await axios.get(url);
        const tokenTransfers = parseHTML(response.data);

        if (tokenTransfers.length === 0 && currentPage === 1) {
            return []; // 如果第一页没有记录，直接返回空数组
        }

        allTokenTransfers = allTokenTransfers.concat(tokenTransfers);

        const parser = new DOMParser();
        const doc = parser.parseFromString(response.data, 'text/html');
        const nextPageLink = doc.querySelector('a.page-link[aria-label="Next"]');
        hasNextPage = !!nextPageLink;
        currentPage++;
    }

    return allTokenTransfers;
}

export async function getTokenBalances(address, provider) {
    const allTokenTransfers = await getAllPages(address);

    if (allTokenTransfers.length === 0) {
        return []; // 如果没有转账记录，返回空数组
    }

    const uniqueTokens = Array.from(new Set(allTokenTransfers.map(t => t.tokenAddress)))
        .map(address => {
            return allTokenTransfers.find(t => t.tokenAddress === address)
        })

    const tokenPromises = uniqueTokens.map(async (token) => {
        const contract = new ethers.Contract(token.tokenAddress, ERC20_ABI, provider)
        try {
            const balance = await contract.balanceOf(address)
            const decimals = await contract.decimals()

            if (balance.gt(0)) {
                return {
                    address: token.tokenAddress,
                    name: token.tokenName,
                    symbol: token.tokenSymbol,
                    balance: ethers.utils.formatUnits(balance, decimals),
                    icon: `/api${token.iconSrc}`
                }
            }
        } catch (error) {
            console.error(`Error fetching token info for ${token.tokenAddress}:`, error)
        }
        return null
    })

    const tokens = await Promise.all(tokenPromises)
    return tokens.filter(token => token !== null)
}


async function getAllNFTPages(address) {
    const url = '/api/nft-transfers.aspx/GetTableData_NftTransfers';
    const requestBody = {
        dataTableModel: {
            start: 0,
            length: 10000,
            Ext: address
        }
    };

    try {
        const response = await axios.post(url, requestBody, {
            headers: {
                'Content-Type': 'application/json'
            }
        });

        const responseData = response.data.d;
        if (!responseData || !responseData.data || responseData.data.length === 0) {
            return [];
        }

        return responseData.data.map(nft => ({
            tokenId: nft.tokenId,
            tokenAddress: nft.tokenAddress,
            tokenName: nft.nftName,
            iconSrc: nft.metadata_image
        }));
    } catch (error) {
        console.error('获取 NFT 数据失败:', error);
        return [];
    }
}

export async function getNFTBalances(address, provider) {
    const allNFTTransfers = await getAllNFTPages(address);

    if (allNFTTransfers.length === 0) {
        return [];
    }

    const uniqueNFTs = Array.from(new Set(allNFTTransfers.map(t => `${t.tokenAddress}-${t.tokenId}`)))
        .map(key => {
            const [tokenAddress, tokenId] = key.split('-');
            return allNFTTransfers.find(t => t.tokenAddress === tokenAddress && t.tokenId === tokenId);
        });

    const nftPromises = uniqueNFTs.map(async (nft) => {
        const contract = new ethers.Contract(nft.tokenAddress, ['function ownerOf(uint256 tokenId) view returns (address)'], provider);
        try {
            const owner = await contract.ownerOf(nft.tokenId);
            if (owner.toLowerCase() === address.toLowerCase()) {
                const imageUrl = await getNFTImageUrl(nft.tokenAddress, nft.tokenId);
                return {
                    tokenId: nft.tokenId,
                    address: nft.tokenAddress,
                    name: nft.tokenName,
                    icon: imageUrl
                };
            }
        } catch (error) {
            console.error(`获取 NFT 信息失败 ${nft.tokenAddress} - ${nft.tokenId}:`, error);
        }
        return null;
    });

    const nfts = await Promise.all(nftPromises);
    return nfts.filter(nft => nft !== null);
}