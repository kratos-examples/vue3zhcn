<script setup lang="ts">
import { ref, reactive, nextTick, onMounted } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'
import { sdk学生管理 } from '../sdk'
import type { T学生信息 } from '../sdk'
import { showErrorDialog } from '../utils/error'
import { showSuccess } from '../utils/message'

const a学生列表 = ref<T学生信息[]>([])
const a日志 = ref<string[]>([])
const ref日志框 = ref<HTMLElement | null>(null)
const b加载中 = ref(false)

// 创建弹窗
const b显示创建弹窗 = ref(false)
const ref创建表单 = ref<FormInstance>()
const v创建表单 = reactive({ v名字: '', v年龄: 18, v班级: '' })
const v创建规则 = reactive<FormRules>({
    v名字: [{ required: true, message: '名字不能为空', trigger: 'blur' }],
})

// 更新弹窗
const b显示更新弹窗 = ref(false)
const ref更新表单 = ref<FormInstance>()
const v更新表单 = reactive({ v编号: '', v名字: '', v年龄: 0, v班级: '' })
const v更新规则 = reactive<FormRules>({
    v名字: [{ required: true, message: '名字不能为空', trigger: 'blur' }],
})

function act记录日志(msg: string) {
    a日志.value.push(`[${new Date().toLocaleTimeString()}] ${msg}`)
    if (a日志.value.length > 50) a日志.value.shift()
    nextTick(() => {
        if (ref日志框.value) ref日志框.value.scrollTop = ref日志框.value.scrollHeight
    })
}

function act打开创建弹窗() {
    v创建表单.v名字 = ''
    v创建表单.v年龄 = 18
    v创建表单.v班级 = ''
    b显示创建弹窗.value = true
    nextTick(() => ref创建表单.value?.clearValidate())
}

async function act创建学生() {
    const valid = await ref创建表单.value?.validate().catch(() => false)
    if (!valid) return
    b加载中.value = true
    try {
        const res = await sdk学生管理.act创建学生({
            v名字: v创建表单.v名字.trim(),
            v年龄: v创建表单.v年龄,
            v班级: v创建表单.v班级.trim(),
        })
        const s = res.s学生
        showSuccess(`创建成功: 编号=${s?.s编号}, 名字=${s?.s名字}`)
        act记录日志(`创建成功: 编号=${s?.s编号}, 名字=${s?.s名字}`)
        await act刷新列表()
        b显示创建弹窗.value = false
    } catch (caught: unknown) {
        showErrorDialog(caught)
        act记录日志(`创建失败: ${caught}`)
    }
    b加载中.value = false
}

function act打开更新弹窗(s: T学生信息) {
    v更新表单.v编号 = s.s编号
    v更新表单.v名字 = s.s名字
    v更新表单.v年龄 = s.s年龄
    v更新表单.v班级 = s.s班级
    b显示更新弹窗.value = true
    nextTick(() => ref更新表单.value?.clearValidate())
}

async function act更新学生() {
    const valid = await ref更新表单.value?.validate().catch(() => false)
    if (!valid) return
    b加载中.value = true
    try {
        const res = await sdk学生管理.act更新学生({
            v编号: v更新表单.v编号,
            v名字: v更新表单.v名字.trim(),
            v年龄: v更新表单.v年龄,
            v班级: v更新表单.v班级.trim(),
        })
        const s = res.s学生
        showSuccess(`更新成功: 编号=${s?.s编号}, 名字=${s?.s名字}`)
        act记录日志(`更新成功: 编号=${s?.s编号}, 名字=${s?.s名字}`)
        await act刷新列表()
        b显示更新弹窗.value = false
    } catch (caught: unknown) {
        showErrorDialog(caught)
        act记录日志(`更新失败: ${caught}`)
    }
    b加载中.value = false
}

async function act刷新列表() {
    b加载中.value = true
    try {
        const res = await sdk学生管理.act学生列表({ v页码: 0, v每页数量: 0 })
        a学生列表.value = res.s学生列表
        act记录日志(`加载 ${res.s学生列表.length} 条学生记录`)
    } catch (caught: unknown) {
        showErrorDialog(caught)
        act记录日志(`列表加载失败: ${caught}`)
    }
    b加载中.value = false
}

async function act查看学生(v编号: string) {
    b加载中.value = true
    try {
        const res = await sdk学生管理.act获取学生({ v编号 })
        const s = res.s学生
        act记录日志(`查询: 编号=${s?.s编号}, 名字=${s?.s名字}, 年龄=${s?.s年龄}, 班级=${s?.s班级}`)
    } catch (caught: unknown) {
        showErrorDialog(caught)
        act记录日志(`查询失败: ${caught}`)
    }
    b加载中.value = false
}

onMounted(() => {
    act刷新列表()
})

async function act删除学生(v编号: string) {
    b加载中.value = true
    try {
        await sdk学生管理.act删除学生({ v编号 })
        showSuccess(`删除成功: 编号=${v编号}`)
        act记录日志(`删除成功: 编号=${v编号}`)
        await act刷新列表()
    } catch (caught: unknown) {
        showErrorDialog(caught)
        act记录日志(`删除失败: ${caught}`)
    }
    b加载中.value = false
}
</script>

<template>
    <div class="demo-section">
        <div class="section-title">
            <h2>学生服务</h2>
            <span class="port">(demo1kratos :8001)</span>
        </div>

        <!-- 学生列表 -->
        <div class="list-card">
            <div class="list-header">
                <h3>学生列表</h3>
                <div class="header-actions">
                    <el-button type="primary" @click="act打开创建弹窗" :loading="b加载中" size="small">创建</el-button>
                    <el-button @click="act刷新列表" :loading="b加载中" size="small">刷新</el-button>
                </div>
            </div>
            <el-table :data="a学生列表" v-loading="b加载中" stripe border v-if="a学生列表.length > 0">
                <el-table-column prop="s编号" label="编号" width="80" align="center" />
                <el-table-column prop="s名字" label="名字" align="center" />
                <el-table-column prop="s年龄" label="年龄" width="80" align="center" />
                <el-table-column prop="s班级" label="班级" align="center" />
                <el-table-column label="操作" width="260" align="center">
                    <template #default="{ row }">
                        <el-button size="small" @click="act查看学生(row.s编号)" :disabled="b加载中">查看</el-button>
                        <el-button size="small" @click="act打开更新弹窗(row)" :disabled="b加载中">编辑</el-button>
                        <el-button size="small" type="danger" @click="act删除学生(row.s编号)" :disabled="b加载中"
                            >删除</el-button
                        >
                    </template>
                </el-table-column>
            </el-table>
            <div v-else class="empty-hint">暂无学生记录，点击创建添加</div>
        </div>

        <!-- 创建弹窗 -->
        <el-dialog v-model="b显示创建弹窗" title="创建学生" width="420" :close-on-click-modal="false">
            <el-form ref="ref创建表单" :model="v创建表单" :rules="v创建规则" label-width="80px">
                <el-form-item label="名字" prop="v名字">
                    <el-input v-model="v创建表单.v名字" placeholder="请输入名字" />
                </el-form-item>
                <el-form-item label="年龄" prop="v年龄">
                    <el-input-number v-model="v创建表单.v年龄" :min="1" :max="200" />
                </el-form-item>
                <el-form-item label="班级" prop="v班级">
                    <el-input v-model="v创建表单.v班级" placeholder="请输入班级" />
                </el-form-item>
            </el-form>
            <template #footer>
                <el-button @click="b显示创建弹窗 = false">取消</el-button>
                <el-button type="primary" @click="act创建学生" :loading="b加载中">提交</el-button>
            </template>
        </el-dialog>

        <!-- 更新弹窗 -->
        <el-dialog
            v-model="b显示更新弹窗"
            :title="`更新学生 (编号: ${v更新表单.v编号})`"
            width="420"
            :close-on-click-modal="false"
        >
            <el-form ref="ref更新表单" :model="v更新表单" :rules="v更新规则" label-width="80px">
                <el-form-item label="名字" prop="v名字">
                    <el-input v-model="v更新表单.v名字" placeholder="请输入名字" />
                </el-form-item>
                <el-form-item label="年龄" prop="v年龄">
                    <el-input-number v-model="v更新表单.v年龄" :min="1" :max="200" />
                </el-form-item>
                <el-form-item label="班级" prop="v班级">
                    <el-input v-model="v更新表单.v班级" placeholder="请输入班级" />
                </el-form-item>
            </el-form>
            <template #footer>
                <el-button @click="b显示更新弹窗 = false">取消</el-button>
                <el-button type="primary" @click="act更新学生" :loading="b加载中">提交</el-button>
            </template>
        </el-dialog>

        <!-- 日志 -->
        <details class="log-section">
            <summary>日志 ({{ a日志.length }})</summary>
            <div class="log-output" ref="ref日志框">
                <div v-for="(line, i) in a日志" :key="i" class="log-line">{{ line }}</div>
            </div>
        </details>
    </div>
</template>

<style scoped>
.demo-section {
    padding: 4px 0;
}

/* Title */
.section-title {
    display: flex;
    align-items: center;
    gap: 10px;
    margin-bottom: 16px;
    padding: 10px 16px;
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    border-radius: 8px;
    color: #fff;
}
.section-title h2 {
    margin: 0;
    font-size: 17px;
}
.port {
    font-size: 13px;
    opacity: 0.8;
    font-weight: normal;
}

/* List card */
.list-card {
    background: #fff;
    border-radius: 8px;
    padding: 16px;
    margin-bottom: 14px;
    border: 1px solid #e4e7ed;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}
.list-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 12px;
}
.list-header h3 {
    margin: 0;
    font-size: 15px;
    color: #303133;
}
.header-actions {
    display: flex;
    gap: 8px;
}
.empty-hint {
    color: #c0c4cc;
    text-align: center;
    padding: 30px;
    font-size: 14px;
}

/* Table */
:deep(.el-table) {
    border: 2px solid #909399;
    border-radius: 6px;
    overflow: hidden;
}
:deep(.el-table th.el-table__cell) {
    background: #f0f2f5;
    color: #303133;
    font-weight: 600;
}
:deep(.el-table td.el-table__cell) {
    border-bottom: 1px dashed #dcdfe6;
}
:deep(.el-table--border .el-table__cell) {
    border-right: 1px dashed #dcdfe6;
}
:deep(.el-table .el-table__cell) {
    padding: 12px 0;
}

/* Logs */
.log-section {
    background: #fff;
    border-radius: 8px;
    padding: 12px 16px;
    border: 1px solid #e4e7ed;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}
.log-section summary {
    cursor: pointer;
    font-size: 13px;
    color: #909399;
    font-weight: 500;
}
.log-output {
    background: #1e1e2e;
    color: #a6e3a1;
    padding: 12px;
    border-radius: 6px;
    font-family: 'Menlo', 'Consolas', monospace;
    font-size: 12px;
    max-height: 200px;
    overflow-y: auto;
    margin-top: 8px;
}
.log-line {
    margin-bottom: 2px;
}
</style>
