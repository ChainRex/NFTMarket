<template>
  <div class="nft-detail" :key="$route.fullPath">
    <el-page-header @back="goBack">
      <template #icon>
        <el-icon class="page-header-icon"><Back /></el-icon>
      </template>
      <template #content>
        <el-skeleton :loading="loading" animated>
          <template #template>
            <el-skeleton-item variant="text" style="width: 150px;" />
          </template>
          <template #default>
            {{ nft.name || 'NFT 详情' }}
          </template>
        </el-skeleton>
      </template>
    </el-page-header>
    
    <el-row :gutter="20" class="mt-4">
      <el-col :span="12">
        <el-skeleton :loading="loading" animated>
          <template #template>
            <div class="nft-image-container">
              <div class="image-preview">
                <el-skeleton-item variant="image" style="width: 100%; height: 100%;" />
              </div>
            </div>
          </template>
          <template #default>
            <div class="nft-image-container">
              <div class="image-preview">
                <img :src="nft.image" class="avatar" :alt="nft.name">
              </div>
            </div>
          </template>
        </el-skeleton>
      </el-col>
      <el-col :span="12">
        <el-skeleton :loading="loading" animated :rows="6">
          <template #default>
            <h2>{{ nft.name }}</h2>
            <p>{{ nft.description }}</p>
            <p>当前所有者: 
              <el-link :href="`https://amoy.polygonscan.com/address/${nft.owner}`" target="_blank">
                {{ shortenAddress(nft.owner) }}
              </el-link>
            </p>
            <p v-if="nft.price">当前价格: {{ formatPrice(nft.price) }} {{ nft.tokenSymbol }}</p>
            
            <template v-if="nft.price">
              <el-button 
                v-if="isCurrentUserSeller" 
                type="info" 
                @click="cancelOrder"
                :loading="isCancelling"
                :disabled="isCancelling"
              >
                <template #icon>
                  <el-icon v-if="!isCancelling"><Close /></el-icon>
                  <el-icon v-else class="is-loading"><Loading /></el-icon>
                </template>
                {{ isCancelling ? '取消中...' : '取消订单' }}
              </el-button>
              <el-button 
                v-else
                type="primary" 
                @click="handleBuyClick"
                :loading="isBuying"
                :disabled="isBuying"
              >
                <template #icon>
                  <el-icon v-if="!isBuying"><ShoppingCart /></el-icon>
                  <el-icon v-else class="is-loading"><Loading /></el-icon>
                </template>
                {{ isBuying ? '购买中...' : '购买' }}
              </el-button>
            </template>
            <template v-else-if="isCurrentUserOwner">
              <el-form :inline="true" class="sell-form">
                <el-form-item class="price-input">
                  <el-input 
                    v-model="sellForm.price" 
                    type="number" 
                    step="1"
                    min="0"
                    placeholder="输入价格"
                  >
                    <template #append>REX</template>
                  </el-input>
                </el-form-item>
                <el-form-item>
                  <el-button 
                    type="primary" 
                    @click="createOrder"
                    :loading="isCreatingOrder"
                    :disabled="isCreatingOrder"
                  >
                    <template #icon>
                      <el-icon v-if="!isCreatingOrder"><Sell /></el-icon>
                      <el-icon v-else class="is-loading"><Loading /></el-icon>
                    </template>
                    {{ isCreatingOrder ? '订单创建中' : '出售' }}
                  </el-button>
                </el-form-item>
              </el-form>
            </template>

            <h3 class="mt-4">属性</h3>
            <el-row :gutter="10">
              <el-col :span="8" v-for="(attr, index) in nft.attributes" :key="index">
                <el-card class="attribute-card">
                  <div class="attribute-type">{{ attr.trait_type }}</div>
                  <div class="attribute-value">{{ attr.value }}</div>
                </el-card>
              </el-col>
            </el-row>
          </template>
        </el-skeleton>
      </el-col>
    </el-row>

    <h3 class="mt-4">活动记录</h3>
    <el-skeleton :loading="historyLoading" animated :rows="5">
      <template #default>
        <el-table :data="transferHistory" style="width: 100%">
          <el-table-column prop="event" label="事件" min-width="30%">
            <template #default="scope">
              <el-icon v-if="scope.row.event === 'Mint'" class="event-icon mint-icon"><Coin /></el-icon>
              <el-icon v-else class="event-icon transfer-icon"><Right /></el-icon>
              {{ scope.row.event }}
            </template>
          </el-table-column>
          <el-table-column prop="from" label="从" min-width="30%">
            <template #default="scope">
              <el-link :href="`https://amoy.polygonscan.com/address/${scope.row.from}`" target="_blank">
                {{ shortenAddress(scope.row.from) }}
              </el-link>
            </template>
          </el-table-column>
          <el-table-column prop="to" label="至" min-width="30%">
            <template #default="scope">
              <el-link :href="`https://amoy.polygonscan.com/address/${scope.row.to}`" target="_blank">
                {{ shortenAddress(scope.row.to) }}
              </el-link>
            </template>
          </el-table-column>
          <el-table-column prop="date" label="日期" min-width="25%">
            <template #default="scope">
              <el-link :href="`https://amoy.polygonscan.com/tx/${scope.row.transactionHash}`" target="_blank">
                {{ formatDate(scope.row.date) }}
              </el-link>
            </template>
          </el-table-column>
        </el-table>
      </template>
    </el-skeleton>

    <!-- 使用的 WalletConnectModal 组件 -->
    <WalletConnectModal :show="showWalletModal" @update:show="showWalletModal = $event" />
  </div>
</template>

<script>
import { ref, computed, onMounted, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useStore } from 'vuex';
import { ethers } from 'ethers';
import { ElSkeleton, ElSkeletonItem } from 'element-plus';
import { getNFTImageUrl, getTokenInfo, getTokenURI } from '../utils/nftUtils';
import { cancelOrder as contractCancelOrder, buyNFT as contractBuyNFT, createOrderWithApprove, initContract } from '../utils/contract';
import NFTMarketAddress from '../contracts/NFTMarket-address.json';
import NFTABI from '../contracts/NFT.json';
import { ElMessage } from 'element-plus';
import { Back, Close, ShoppingCart, Sell, Coin, Right, Loading } from '@element-plus/icons-vue';
import WalletConnectModal from './WalletConnectModal.vue';
import { getProvider } from '../utils/contract';
import { handleGlobalError } from '../utils/errorHandler';
export default {
  components: {
    Back,
    Close,
    ShoppingCart,
    Sell,
    WalletConnectModal,
    ElSkeleton,
    ElSkeletonItem,
    Coin,
    Right,
    Loading,
  },
  setup() {
    const route = useRoute();
    const router = useRouter();
    const store = useStore();
    const nft = ref({});
    const transferHistory = ref([]);
    const loading = ref(true);
    const historyLoading = ref(true);
    const isRefreshing = ref(false); // 新增：用于防止重复刷新的标志

    const isWalletConnected = computed(() => store.state.isWalletConnected);
    const currentUserAddress = computed(() => store.state.currentUserAddress);

    const isCurrentUserSeller = computed(() => {
      return nft.value.owner && currentUserAddress.value && 
             nft.value.owner.toLowerCase() === currentUserAddress.value.toLowerCase();
    });

    const sellForm = ref({
      price: ''
    });

    const isCurrentUserOwner = computed(() => {
      return nft.value.owner && currentUserAddress.value && 
             nft.value.owner.toLowerCase() === currentUserAddress.value.toLowerCase();
    });

    const isCancelling = ref(false);
    const isBuying = ref(false);
    const isCreatingOrder = ref(false);
    
    const cancelOrder = async () => {
      if (isCancelling.value) return;
      isCancelling.value = true;
      try {
        if (!isWalletConnected.value) {
          ElMessage.warning('请先连接钱包');
          showWalletModal.value = true;
          return;
        }

        const { collectionAddress, tokenId } = route.params;
        const rawOrders = await store.dispatch("fetchOrders");
        const orderIndex = rawOrders.findIndex(o => 
          o.nft === collectionAddress && 
          ethers.BigNumber.from(o.tokenId).eq(ethers.BigNumber.from(tokenId)) &&
          o.status._hex === '0x00'
        );

        if (orderIndex === -1) {
          throw new Error('找不到当前 NFT 的活跃订单');
        }
        
        const currentUserAddress = store.state.currentUserAddress;
        if (!currentUserAddress) {
          throw new Error('无法获取当前用户地址');
        }

        if (rawOrders[orderIndex].seller.toLowerCase() !== currentUserAddress.toLowerCase()) {
          throw new Error('您不是该订单的创建者，无法取消订单');
        }

        await initContract(true);
        
        ElMessage.info('请在钱包中确认交易...');
        await contractCancelOrder(orderIndex);
        ElMessage.success('订单已成功取消');
        await fetchNFTDetails();
      } catch (error) {
        console.error('取消订单失败:', error);
        ElMessage.error('取消订单失败: ' + error.message);
      } finally {
        isCancelling.value = false;
      }
    };

    const buyNFT = async () => {
      if (isBuying.value) return;
      isBuying.value = true;
      try {
        if (!isWalletConnected.value) {
          ElMessage.warning('请先连接钱包');
          showWalletModal.value = true;
          return;
        }

        const { collectionAddress, tokenId } = route.params;
        const rawOrders = await store.dispatch("fetchOrders");
        const originalOrder = rawOrders.find(o => 
          o.nft === collectionAddress && 
          ethers.BigNumber.from(o.tokenId).eq(ethers.BigNumber.from(tokenId)) &&
          o.status._hex === '0x00'
        );

        if (!originalOrder) {
          throw new Error('找不到当前 NFT 的活跃订单');
        }

        console.log('找到的订单:', originalOrder);

        const order = { ...originalOrder };

        if (!order.token || !order.tokenName) {
          const tokenInfo = await getTokenInfo(order.token);
          order.tokenName = tokenInfo.name;
        }

        const deadline = Math.floor(Date.now() / 1000) + 3600; // 1小时后过
        const { v, r, s } = await getSignature(order, deadline);
        
        const orderIndex = rawOrders.indexOf(originalOrder);

        console.log('准备购买 NFT，参数:', {
          orderIndex,
          deadline,
          v,
          r,
          s
        });

        // 初始化合约，确保使用钱包
        await initContract(true);

        ElMessage.info('请在钱包中确认交易...');
        const tx = await contractBuyNFT(orderIndex, deadline, v, r, s);
        if (!tx) {
          throw new Error('购买交易失败');
        }
        ElMessage.success('NFT 购买交易已提交，等待确认...');
        
        const receipt = await tx.wait();
        console.log('交易确认:', receipt);
        
        ElMessage.success('NFT 购买成功，正在更新详情...');
        
        await fetchNFTDetails();
        ElMessage.success('NFT 详情已更新');
      } catch (error) {
        console.error('购买 NFT 失败:', error);
        ElMessage.error('购买 NFT 失败: ' + error.message);
      } finally {
        isBuying.value = false;
      }
    };

    const createOrder = async () => {
      if (isCreatingOrder.value) return;
      isCreatingOrder.value = true;
      try {
        if (!sellForm.value.price) {
          ElMessage.warning('请填写价格');
          return;
        }
        if (parseFloat(sellForm.value.price) <= 0) {
          ElMessage.warning('价格必须大于 0');
          return;
        }
        try {
          const { collectionAddress, tokenId } = route.params;
          ElMessage.info('正在创建订单，请稍候...');

          // 重新初始化合约
          await initContract(true);

          // 获取最新的 provider 和 signer
          const provider = await getProvider();
          const signer = provider.getSigner();

          // 确保有签名者
          if (!signer) {
            throw new Error('无法获取签名者，请确保钱包已连接');
          }

          ElMessage.info('请在钱包中确认交易...');
          const success = await createOrderWithApprove(
            collectionAddress,
            tokenId,
            store.state.rexContractAddress,
            sellForm.value.price.toString(),
            signer
          );
          
          if (success) {
            ElMessage.success('订单创建成功');
            await fetchNFTDetails();
          } else {
            ElMessage.error('订单创建失败');
          }
        } catch (error) {
          console.error('创建订单失败:', error);
          ElMessage.error('创建订单失败: ' + error.message);
        }
      } catch (error) {
        console.error('创建订单失败:', error);
        ElMessage.error('创建订单失败: ' + error.message);
      } finally {
        isCreatingOrder.value = false;
      }
    };

    const fetchNFTDetails = async () => {
      if (isRefreshing.value) return; // 如果正在刷新，则直接返回
      isRefreshing.value = true; // 设置刷新标志
      try {
        loading.value = true;
        const { collectionAddress, tokenId } = route.params;
        
        const [
          currentOwner,
          rawOrders,
          tokenURI,
          imageUrl
        ] = await Promise.all([
          getNFTOwner(collectionAddress, tokenId),
          store.dispatch("fetchOrders", { force: true }),
          getTokenURI(collectionAddress, tokenId), 
          getNFTImageUrl(collectionAddress, tokenId)
        ]);
        
        console.log('当前 NFT 所有者:', currentOwner);
        
        const activeOrder = rawOrders.find(o => 
          o.nft === collectionAddress && 
          ethers.BigNumber.from(o.tokenId).eq(ethers.BigNumber.from(tokenId)) &&
          o.status._hex === '0x00'  // 只查找未售出的订单
        );
        
        console.log('Token URI:', tokenURI);
        const metadata = await fetchMetadata(tokenURI);
        console.log('Metadata:', metadata);
        console.log('Image URL:', imageUrl);

        nft.value = {
          ...metadata,
          image: imageUrl,
          owner: currentOwner,
          collectionAddress,
          tokenId
        };

        if (activeOrder) {
          const tokenInfo = await getTokenInfo(activeOrder.token);
          nft.value = {
            ...nft.value,
            price: activeOrder.price,
            tokenSymbol: tokenInfo.symbol,
            tokenName: tokenInfo.name,
            token: activeOrder.token,
            orderId: activeOrder.id,
            seller: activeOrder.seller
          };
        }

        await fetchTransferHistory(collectionAddress, tokenId);
        
        console.log('更新后的 NFT 详情:', nft.value);
      } catch (error) {
        console.error('获取 NFT 详情失败:', error);
        ElMessage.error('获取 NFT 详情失败: ' + error.message);
      } finally {
        loading.value = false;
        isRefreshing.value = false; // 重置刷新标志
      }
    };

    const fetchMetadata = async (tokenURI) => {
      try {
        // 定义 IPFS 网关列表
        const IPFS_GATEWAYS = [
          'https://gateway.pinata.cloud/ipfs/',
        ];

        let metadata;
        for (const gateway of IPFS_GATEWAYS) {
          try {
            // 如果 tokenURI 是 IPFS URL，将其转换为 HTTP URL
            const url = tokenURI.replace('ipfs://', gateway);
            console.log('Trying to fetch metadata from:', url);
            const response = await fetch(url);
            if (!response.ok) {
              throw new Error(`HTTP error! status: ${response.status}`);
            }
            metadata = await response.json();
            console.log('Metadata fetched successfully:', metadata);
            break; // 如果成功取元数据，跳出循环
          } catch (error) {
            console.error(`Failed to fetch metadata from ${gateway}:`, error);
          }
        }

        if (!metadata) {
          throw new Error('Failed to fetch metadata from all gateways');
        }

        return metadata;
      } catch (error) {
        console.error('获取或解析元数据失败:', error);
        return {
          name: 'Unknown',
          description: 'Metadata unavailable',
          attributes: []
        };
      }
    };

    const fetchTransferHistory = async (nftAddress, tokenId) => {
      try {
        historyLoading.value = true;
        const provider = await getProvider();
        const nftContract = new ethers.Contract(nftAddress, NFTABI.abi, provider);

        const transferFilter = nftContract.filters.Transfer(null, null, tokenId);
        const transferEvents = await nftContract.queryFilter(transferFilter);

        const allEvents = transferEvents.sort((a, b) => b.blockNumber - a.blockNumber);

        const historyPromises = allEvents.map(async (event) => {
          const block = await provider.getBlock(event.blockNumber);
          return {
            event: event.args.from === ethers.constants.AddressZero ? 'Mint' : 'Transfer',
            from: event.args.from,
            to: event.args.to,
            date: new Date(block.timestamp * 1000),
            transactionHash: event.transactionHash
          };
        });

        transferHistory.value = await Promise.all(historyPromises);
      } catch (error) {
        console.error('获取转移历史失败:', error);
        ElMessage.error('获取转移历史失败: ' + error.message);
        handleGlobalError(error);
      } finally {
        historyLoading.value = false;
      }
    };

    const formatPrice = (price) => {
      if (!price) return 'N/A';
      try {
        return ethers.utils.formatEther(price);
      } catch (error) {
        console.error('格式化价格失败:', error, price);
        return 'Error';
      }
    };

    const formatDate = (date) => {
      return date.toLocaleString();
    };

    const goBack = () => {
      router.back();
    };

    const shortenAddress = (address) => {
      if (!address) return 'N/A';
      return `${address.slice(0, 6)}...`;
    };

    const getSignature = async (order, deadline) => {
      console.log('开始获取签名，订单信息：', order);
      
      if (!order || !order.token || !order.tokenName) {
        console.error('订单信息不完整:', order);
        throw new Error('订单信息不完整，无法生成签名');
      }

      const provider = await getProvider();
      const signer = provider.getSigner();
      
      const domain = {
        name: order.tokenName,
        version: '1',
        chainId: 80002, // Polygon Mumbai 测试网的 chainId
        verifyingContract: order.token
      };

      console.log('Domain 信息:', domain);

      const types = {
        Permit: [
          { name: 'owner', type: 'address' },
          { name: 'spender', type: 'address' },
          { name: 'value', type: 'uint256' },
          { name: 'nonce', type: 'uint256' },
          { name: 'deadline', type: 'uint256' }
        ]
      };

      try {
        const tokenContract = new ethers.Contract(
          order.token,
          ['function nonces(address owner) view returns (uint256)'],
          provider
        );

        const signerAddress = await signer.getAddress();
        console.log('签名者地址:', signerAddress);

        const nonce = await tokenContract.nonces(signerAddress);
        console.log('Nonce:', nonce.toString());

        const value = {
          owner: signerAddress,
          spender: NFTMarketAddress.address,
          value: order.price,
          nonce: nonce,
          deadline: deadline
        };

        console.log('签名数据:', value);

        const signature = await signer._signTypedData(domain, types, value);
        console.log('签名成功:', signature);
        return ethers.utils.splitSignature(signature);
      } catch (error) {
        console.error("签名过程中出错:", error);
        throw new Error("签名失: " + error.message);
      }
    };

    // 增函数：直接从 NFT 合约获取所有者信息
    const getNFTOwner = async (nftAddress, tokenId) => {
      const provider = await getProvider();
      const nftContract = new ethers.Contract(nftAddress, ['function ownerOf(uint256 tokenId) view returns (address)'], provider);
      return await nftContract.ownerOf(tokenId);
    };

    const showWalletModal = ref(false);

    const handleBuyClick = async () => {
      if (!isWalletConnected.value) {
        showWalletModal.value = true;
      } else {
        await buyNFT();
      }
    };

    onMounted(async () => {
      await initContract();
      await fetchNFTDetails();
    });

    // 监听钱包连接状态变化
    watch(() => store.state.isWalletConnected, async (newValue) => {
      if (newValue) {
        await fetchNFTDetails();
      }
    });

    // 使用 watch 替代 watchEffect 来监听路由参数变化
    watch(() => [route.params.collectionAddress, route.params.tokenId], async (newValue, oldValue) => {
      if (newValue[0] && newValue[1] && (newValue[0] !== oldValue[0] || newValue[1] !== oldValue[1])) {
        await fetchNFTDetails();
      }
    });

    return {
      nft,
      transferHistory,
      formatPrice,
      formatDate,
      goBack,
      shortenAddress,
      isCurrentUserSeller,
      cancelOrder,
      buyNFT,
      isCurrentUserOwner,
      sellForm,
      createOrder,
      isWalletConnected,
      showWalletModal,
      handleBuyClick,
      loading,
      historyLoading,
      isCancelling,
      isBuying,
      isCreatingOrder,
    };
  }
};
</script>

<style scoped>
.nft-detail {
  max-width: 1200px;
  margin: 0 auto;
  padding: 20px;
}

.nft-image-container {
  width: 100%;
  padding-top: 100%; /* 创建一个正方形容器 */
  position: relative;
  overflow: hidden;
  border: 1px solid #d9d9d9;
  border-radius: 6px;
}

.image-preview {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  display: flex;
  justify-content: center;
  align-items: center;
  background-color: #ffffff;
}

.avatar {
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
}

.mt-4 {
  margin-top: 1rem;
}

.attribute-card {
  margin-bottom: 10px;
  text-align: center;
}

.attribute-type {
  font-weight: bold;
  margin-bottom: 5px;
}

.attribute-value {
  color: #606266;
}

.el-table {
  margin-top: 1rem;
  width: 100%;
}

.el-table .el-table__cell {
  padding: 8px;
}

@media (max-width: 768px) {
  .el-table {
    font-size: 12px;
  }
  
  .el-table .el-table__cell {
    padding: 5px;
  }
}

.el-button {
  margin-top: 20px;
}

.el-button .el-icon {
  margin-right: 5px;
  vertical-align: middle;
}

.page-header-icon {
  margin-right: 8px;
}

.el-link {
  text-decoration: none;
}

.el-link:hover {
  text-decoration: none;
}

.wallet-modal-content {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.metamask-button {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  padding: 10px;
  margin-top: 20px;
  background-color: #ffffff;
  color: black;
  border: none;
  border-radius: 12px;
  font-weight: 600;
  cursor: pointer;
  transition: background-color 0.2s ease;
}

.metamask-button:hover {
  background-color: #e5e8eb;
}

.metamask-button img {
  width: 24px;
  height: 24px;
  margin-right: 10px;
}

.close-button {
  margin-top: 20px;
  padding: 10px 20px;
  background-color: #e5e8eb;
  border: none;
  border-radius: 12px;
  cursor: pointer;
  transition: background-color 0.2s ease;
}

.close-button:hover {
  background-color: #d0d5da;
}

.skeleton-header {
  display: flex;
  align-items: center;
  height: 40px;
  margin-bottom: 20px;
}

.sell-form {
  display: flex;
  align-items: center;
  margin-top: 20px;
}

.sell-form .el-form-item {
  margin-bottom: 0;
  margin-right: 10px;
}

.sell-form .price-input {
  flex-grow: 0; /* 修改： 1 改为 0 */
  width: 200px; /* 新增：设置固定宽度 */
}

.sell-form .el-button {
  margin-top: 0;
}

/* 添加以下新样式 */
.sell-form .el-input-number__decrease,
.sell-form .el-input-number__increase {
  display: none;
}

.sell-form .el-input-number .el-input__inner {
  -moz-appearance: textfield;
}

.sell-form .el-input-number .el-input__inner::-webkit-outer-spin-button,
.sell-form .el-input-number .el-input__inner::-webkit-inner-spin-button {
  -webkit-appearance: none;
  margin: 0;
}

.event-icon {
  margin-right: 5px;
  vertical-align: middle;
}

.mint-icon {
  color: #E6A23C; /* 金色 */
}

.transfer-icon {
  color: #409EFF; /* 蓝色 */
}

.el-button .is-loading {
  animation: rotating 2s linear infinite;
}

@keyframes rotating {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}
</style>