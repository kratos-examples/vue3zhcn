import { ElMessage } from 'element-plus'

export function showSuccess(msg: string) {
    ElMessage.success({ message: msg, duration: 1500 })
}
