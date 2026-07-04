import { ElMessage } from 'element-plus'

export function showSuccess(msg: string) {
    ElMessage.success({ message: msg, duration: 1500 })
}

export function showWarning(msg: string) {
    ElMessage.warning({ message: msg, duration: 2000 })
}
