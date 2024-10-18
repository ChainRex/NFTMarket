<template>
  <div class="collection-detail">
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
            {{ collection.name || 'NFT 系列' }}
          </template>
        </el-skeleton>
      </template>
    </el-page-header>
    
    <!-- 集合头部信息 -->
    <el-skeleton :loading="loading" animated>
      <template #template>
        <el-row class="collection-header">
          <el-col :span="4">
            <el-skeleton-item variant="circle" style="width: 100px; height: 100px;" />
          </el-col>
          <el-col :span="20">
            <el-skeleton-item variant="h1" style="width: 50%;" />
            <el-skeleton-item variant="text" style="width: 60%;" />
            <el-skeleton-item variant="text" style="width: 40%;" />
          </el-col>
        </el-row>
      </template>
      <template #default>
        <el-row class="collection-header">
          <el-col :span="4">
            <el-avatar :size="100" :src="collection.iconUrl"></el-avatar>
          </el-col>
          <el-col :span="20">
            <h1>
              <a :href="getExplorerUrl('token', collection.address)" target="_blank" rel="noopener noreferrer">
                {{ collection.name }}
              </a>
            </h1>
            <p class="contract-address">
              合约地址: 
              <a :href="getExplorerUrl('address', collection.address)" target="_blank" rel="noopener noreferrer">
                {{ collection.address }}
              </a>
              <el-button class="copy-button" @click="copyAddress(collection.address)" type="text">
                <el-icon><DocumentCopy /></el-icon>
              </el-button>
            </p>
            <p v-if="collection.floorPrice && collection.tokenSymbol">地板价: {{ formatPrice(collection.floorPrice) }} {{ collection.tokenSymbol }}</p>
          </el-col>
        </el-row>
      </template>
    </el-skeleton>

    <!-- NFT 列表 -->
    <el-row :gutter="20">
      <template v-if="loading">
        <el-col :span="6" v-for="i in 8" :key="i">
          <el-card :body-style="{ padding: '0px' }" shadow="hover">
            <el-skeleton :loading="loading" animated>
              <template #template>
                <el-skeleton-item variant="image" style="width: 100%; height: 200px;" />
                <div style="padding: 14px;">
                  <el-skeleton-item variant="h3" style="width: 50%;" />
                  <div style="display: flex; align-items: center; justify-content: space-between; margin-top: 13px;">
                    <el-skeleton-item variant="text" style="width: 60%;" />
                  </div>
                </div>
              </template>
            </el-skeleton>
          </el-card>
        </el-col>
      </template>
      
      <template v-else>
        <el-col :span="6" v-for="nft in nfts" :key="nft.id">
          <router-link :to="`/nft/${collection.address}/${nft.id}`" class="nft-link">
            <el-card :body-style="{ padding: '0px' }" shadow="hover">
              <img :src="nft.imageUrl" class="image" :alt="nft.name">
              <div style="padding: 14px;">
                <span>{{ collection.name }} #{{ nft.id }} - {{ nft.name }}</span>
                <div class="bottom">
                  <span v-if="nft.price">{{ formatPrice(nft.price) }} {{ collection.tokenSymbol }}</span>
                  <span v-else>&nbsp;</span>
                </div>
              </div>
            </el-card>
          </router-link>
        </el-col>
      </template>
    </el-row>
  </div>
</template>

<script>
import { ref, onMounted, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useStore } from 'vuex';
import { ethers } from 'ethers';
import { ElSkeleton, ElSkeletonItem, ElMessage } from 'element-plus';
import { Back, DocumentCopy } from '@element-plus/icons-vue';
import { getNFTImageUrl, getTokenInfo, getNFTName, getNFTTokenIconURI, getIPFSUrl } from '../utils/nftUtils';
import NFTABI from '../contracts/NFT.json';
import { getProvider } from '../utils/contract';

export default {
  components: {
    ElSkeleton,
    ElSkeletonItem,
    Back,
    DocumentCopy,
  },
  setup() {
    const route = useRoute();
    const router = useRouter();
    const store = useStore();
    const collection = ref({});
    const nfts = ref([]);
    const loading = ref(true);

    const fetchCollectionDetails = async () => {
      try {
        loading.value = true;
        const collectionAddress = route.params.address;
        const provider = await getProvider();
        const nftContract = new ethers.Contract(collectionAddress, NFTABI.abi, provider);
        
        const [rawOrders, name, tokenIconURI, tokenURIList] = await Promise.all([
          store.dispatch("fetchOrders"),
          getNFTName(collectionAddress),
          getNFTTokenIconURI(collectionAddress),
          nftContract.getTokenURIList()
        ]);

        const iconUrl = getIPFSUrl(tokenIconURI);

        collection.value = {
          address: collectionAddress,
          name,
          iconUrl,
          floorPrice: ethers.constants.MaxUint256,
          tokenSymbol: ''
        };

        const nftPromises = tokenURIList.map(async (tokenURI, index) => {
          const [metadata, imageUrl] = await Promise.all([
            fetchMetadata(tokenURI),
            getNFTImageUrl(collectionAddress, index)
          ]);

          const order = rawOrders.find(o => 
            o.nft === collectionAddress && 
            ethers.BigNumber.from(o.tokenId).eq(ethers.BigNumber.from(index)) &&
            o.status._hex === '0x00'
          );

          if (order && order.price) {
            const price = ethers.BigNumber.from(order.price);
            if (price.lt(collection.value.floorPrice)) {
              collection.value.floorPrice = price;
            }
            if (!collection.value.tokenSymbol) {
              const tokenInfo = await getTokenInfo(order.token);
              collection.value.tokenSymbol = tokenInfo.symbol;
            }
          }

          return {
            id: index,
            name: metadata.name,
            imageUrl: imageUrl,
            price: order ? order.price : null
          };
        });

        nfts.value = await Promise.all(nftPromises);

        if (collection.value.floorPrice.eq(ethers.constants.MaxUint256)) {
          collection.value.floorPrice = null;
        }

      } catch (error) {
        console.error('获取集合详情失败:', error);
      } finally {
        loading.value = false;
      }
    };

    const fetchMetadata = async (tokenURI) => {
      try {
        const response = await fetch(getIPFSUrl(tokenURI));
        if (!response.ok) {
          throw new Error(`HTTP error! status: ${response.status}`);
        }
        return await response.json();
      } catch (error) {
        console.error('获取元数据失败:', error);
        return { name: 'Unknown' };
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

    const goBack = () => {
      router.push('/');
    };

    const getExplorerUrl = (type, address) => {
      const baseUrl = 'https://amoy.polygonscan.com';
      switch (type) {
        case 'token':
          return `${baseUrl}/token/${address}`;
        case 'address':
          return `${baseUrl}/address/${address}`;
        default:
          return baseUrl;
      }
    };

    const copyAddress = (address) => {
      navigator.clipboard.writeText(address).then(() => {
        ElMessage.success('地址已复制到剪贴板');
      }).catch(err => {
        console.error('复制失败:', err);
        ElMessage.error('复制地址失败');
      });
    };

    onMounted(fetchCollectionDetails);

    watch(
      () => route.params.address,
      (newAddress, oldAddress) => {
        if (newAddress !== oldAddress) {
          fetchCollectionDetails();
        }
      }
    );

    return {
      collection,
      nfts,
      loading,
      formatPrice,
      goBack,
      getExplorerUrl,
      copyAddress
    };
  }
};
</script>

<style scoped>
.collection-detail {
  max-width: 1200px;
  margin: 0 auto;
  padding: 20px;
}

.collection-header {
  margin: 20px 0;
}

.image {
  width: 100%;
  display: block;
}

.bottom {
  margin-top: 13px;
  line-height: 12px;
}

.el-card {
  margin-bottom: 20px;
}

.nft-link {
  text-decoration: none;
  color: inherit;
}

a {
  color: #409EFF;
  text-decoration: none;
}

a:hover {
  text-decoration: none; /* 移除悬浮时的下划线 */
}

.page-header-icon {
  margin-right: 8px;
}

.contract-address {
  display: flex;
  align-items: center;
}

.copy-button {
  margin-left: 8px;
  padding: 2px;
}

.copy-button .el-icon {
  font-size: 16px;
}
</style>