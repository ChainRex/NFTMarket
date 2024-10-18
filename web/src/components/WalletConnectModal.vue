<template>
  <div v-if="show" class="wallet-modal">
    <div class="wallet-modal-content">
      <h3>连接钱包</h3>
      <button @click="connectMetamask" class="metamask-button">
        <img src="../assets/metamask-logo.svg" alt="MetaMask" />
        MetaMask
      </button>
      <button @click="close" class="close-button">关闭</button>
    </div>
  </div>
</template>

<script>
import { useStore } from 'vuex'

export default {
  name: 'WalletConnectModal',
  props: {
    show: {
      type: Boolean,
      required: true
    }
  },
  emits: ['update:show'],
  setup(props, { emit }) {
    const store = useStore()

    const connectMetamask = async () => {
      const success = await store.dispatch('connectWallet')
      if (success) {
        console.log('钱包连接成功')
        emit('update:show', false)
      } else {
        console.error('钱包连接失败')
      }
    }

    const close = () => {
      emit('update:show', false)
    }

    return {
      connectMetamask,
      close
    }
  }
}
</script>

<style scoped>
.wallet-modal {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
}

.wallet-modal-content {
  background-color: white;
  padding: 20px;
  border-radius: 12px;
  width: 300px;
  text-align: center;
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
</style>