import { ref } from 'vue';
import { ElMessage } from 'element-plus';

export const globalError = ref(null);

export function handleGlobalError(error) {
    console.error('Global error:', error);
    if (error.message && error.message.includes('Internal JSON-RPC error')) {
        globalError.value = {
            type: 'rpc',
            message: '当前 RPC 节点可能不稳定，请重试'
        };
        ElMessage.error('当前 RPC 节点可能不稳定，请重试');
    } else {
        globalError.value = {
            type: 'general',
            message: error.message
        };
        ElMessage.error(error.message);
    }
}

export function clearGlobalError() {
    globalError.value = null;
}