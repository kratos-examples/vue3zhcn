import type { AxiosError } from 'axios'
import { isAxiosError } from 'axios'
import { ElMessageBox } from 'element-plus'
import { h } from 'vue'

// Kratos error response structure
interface KratosErrorData {
    reason: string
    code: number
    message: string
    metadata: Record<string, string | undefined>
}

// Parsed error info
export interface ErrorInfo {
    httpCode: number
    reason: string
    message: string
}

// Parse Kratos error from AxiosError
export function parseError(caught: unknown): ErrorInfo {
    if (isAxiosError(caught)) {
        const axiosErr = caught as AxiosError
        if (axiosErr.code === 'ERR_NETWORK') {
            return { httpCode: 0, reason: 'ERR_NETWORK', message: 'Network error, check connection' }
        }
        if (axiosErr.response?.data) {
            const data = axiosErr.response.data as KratosErrorData
            return {
                httpCode: data.code ?? axiosErr.response.status ?? 500,
                reason: data.reason ?? 'UNKNOWN',
                message: data.message ?? axiosErr.message,
            }
        }
    }
    return { httpCode: 500, reason: 'UNKNOWN', message: String(caught) }
}

// Show error in a dialog
export function showErrorDialog(caught: unknown) {
    const info = parseError(caught)
    ElMessageBox({
        title: info.httpCode === 0 ? 'Network Error' : `Error (${info.httpCode})`,
        message: h('div', { style: 'font-size:14px' }, [
            h('p', { style: 'margin:0 0 8px 0;color:#606266' }, [h('b', 'Reason: '), info.reason]),
            h('p', { style: 'margin:0;color:#909399;word-break:break-all' }, info.message),
        ]),
        confirmButtonText: 'OK',
        type: 'warning',
    })
}
