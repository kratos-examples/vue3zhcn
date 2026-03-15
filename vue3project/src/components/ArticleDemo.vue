<script setup lang="ts">
import { ref, reactive, nextTick, onMounted } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'
import { demo2Transport } from '../api/transport'
import { ArticleServiceClient } from '../rpc/demo2/article/article.client'
import {
    CreateArticleRequest,
    UpdateArticleRequest,
    DeleteArticleRequest,
    GetArticleRequest,
    ListArticlesRequest,
} from '../rpc/demo2/article/article'
import type { ArticleInfo } from '../rpc/demo2/article/article'
import { showErrorDialog } from '../utils/error'
import { showSuccess } from '../utils/message'

const client = new ArticleServiceClient(demo2Transport)

const articles = ref<ArticleInfo[]>([])
const logs = ref<string[]>([])
const logBox = ref<HTMLElement | null>(null)
const loading = ref(false)

// create dialog
const showCreateDialog = ref(false)
const createFormRef = ref<FormInstance>()
const createForm = reactive({ title: '', content: '', studentId: '' })
const createRules = reactive<FormRules>({
    title: [{ required: true, message: 'Title is required', trigger: 'blur' }],
})

// update dialog
const showUpdateDialog = ref(false)
const updateFormRef = ref<FormInstance>()
const updateForm = reactive({ id: '', title: '', content: '', studentId: '' })
const updateRules = reactive<FormRules>({
    title: [{ required: true, message: 'Title is required', trigger: 'blur' }],
})

function log(msg: string) {
    logs.value.push(`[${new Date().toLocaleTimeString()}] ${msg}`)
    if (logs.value.length > 50) logs.value.shift()
    nextTick(() => {
        if (logBox.value) logBox.value.scrollTop = logBox.value.scrollHeight
    })
}

function openCreate() {
    createForm.title = ''
    createForm.content = ''
    createForm.studentId = ''
    showCreateDialog.value = true
    nextTick(() => createFormRef.value?.clearValidate())
}

async function doCreate() {
    const valid = await createFormRef.value?.validate().catch(() => false)
    if (!valid) return
    loading.value = true
    try {
        const request = CreateArticleRequest.create({
            title: createForm.title.trim(),
            content: createForm.content.trim(),
            studentId: createForm.studentId.trim() || '0',
        })
        const response = await client.createArticle(request, {})
        const a = response.data.article
        showSuccess(`Created: id=${a?.id}, title=${a?.title}`)
        log(`Created: id=${a?.id}, title=${a?.title}`)
        await doList()
        showCreateDialog.value = false
    } catch (caught: unknown) {
        showErrorDialog(caught)
        log(`Create FAIL: ${caught}`)
    }
    loading.value = false
}

function openUpdate(a: ArticleInfo) {
    updateForm.id = a.id
    updateForm.title = a.title
    updateForm.content = a.content
    updateForm.studentId = a.studentId
    showUpdateDialog.value = true
    nextTick(() => updateFormRef.value?.clearValidate())
}

async function doUpdate() {
    const valid = await updateFormRef.value?.validate().catch(() => false)
    if (!valid) return
    loading.value = true
    try {
        const request = UpdateArticleRequest.create({
            id: updateForm.id,
            title: updateForm.title.trim(),
            content: updateForm.content.trim(),
            studentId: updateForm.studentId.trim() || '0',
        })
        const response = await client.updateArticle(request, {})
        const a = response.data.article
        showSuccess(`Updated: id=${a?.id}, title=${a?.title}`)
        log(`Updated: id=${a?.id}, title=${a?.title}`)
        await doList()
        showUpdateDialog.value = false
    } catch (caught: unknown) {
        showErrorDialog(caught)
        log(`Update FAIL: ${caught}`)
    }
    loading.value = false
}

async function doList() {
    loading.value = true
    try {
        const request = ListArticlesRequest.create({})
        const response = await client.listArticles(request, {})
        articles.value = response.data.articles
        log(`Loaded ${response.data.articles.length} articles`)
    } catch (caught: unknown) {
        showErrorDialog(caught)
        log(`List FAIL: ${caught}`)
    }
    loading.value = false
}

async function doGet(id: string) {
    loading.value = true
    try {
        const request = GetArticleRequest.create({ id })
        const response = await client.getArticle(request, {})
        const a = response.data.article
        log(`Get: id=${a?.id}, title=${a?.title}, content=${a?.content}, studentId=${a?.studentId}`)
    } catch (caught: unknown) {
        showErrorDialog(caught)
        log(`Get FAIL: ${caught}`)
    }
    loading.value = false
}

onMounted(() => {
    doList()
})

async function doDelete(id: string) {
    loading.value = true
    try {
        const request = DeleteArticleRequest.create({ id })
        await client.deleteArticle(request, {})
        showSuccess(`Deleted: id=${id}`)
        log(`Deleted: id=${id}`)
        await doList()
    } catch (caught: unknown) {
        showErrorDialog(caught)
        log(`Delete FAIL: ${caught}`)
    }
    loading.value = false
}
</script>

<template>
    <div class="demo-section">
        <div class="section-title">
            <h2>ArticleService</h2>
            <span class="port">(demo2kratos :8002)</span>
        </div>

        <!-- Article List -->
        <div class="list-card">
            <div class="list-header">
                <h3>Articles</h3>
                <div class="header-actions">
                    <el-button type="primary" @click="openCreate" :loading="loading" size="small">Create</el-button>
                    <el-button @click="doList" :loading="loading" size="small">Refresh</el-button>
                </div>
            </div>
            <el-table :data="articles" v-loading="loading" stripe border v-if="articles.length > 0">
                <el-table-column prop="id" label="ID" width="80" align="center" />
                <el-table-column prop="title" label="Title" align="center" />
                <el-table-column prop="content" label="Content" align="center" show-overflow-tooltip />
                <el-table-column label="Student" width="100" align="center">
                    <template #default="{ row }">
                        {{ row.studentId !== '0' ? row.studentId : '-' }}
                    </template>
                </el-table-column>
                <el-table-column label="Actions" width="260" align="center">
                    <template #default="{ row }">
                        <el-button size="small" @click="doGet(row.id)" :disabled="loading">Select</el-button>
                        <el-button size="small" @click="openUpdate(row)" :disabled="loading">Update</el-button>
                        <el-button size="small" type="danger" @click="doDelete(row.id)" :disabled="loading"
                            >Delete</el-button
                        >
                    </template>
                </el-table-column>
            </el-table>
            <div v-else class="empty-hint">No articles yet. Click Create to add one.</div>
        </div>

        <!-- Create Dialog -->
        <el-dialog v-model="showCreateDialog" title="Create Article" width="420" :close-on-click-modal="false">
            <el-form ref="createFormRef" :model="createForm" :rules="createRules" label-width="90px">
                <el-form-item label="Title" prop="title">
                    <el-input v-model="createForm.title" placeholder="Title" />
                </el-form-item>
                <el-form-item label="Content" prop="content">
                    <el-input v-model="createForm.content" type="textarea" placeholder="Content" :rows="3" />
                </el-form-item>
                <el-form-item label="Student ID" prop="studentId">
                    <el-input v-model="createForm.studentId" placeholder="Student ID (optional)" />
                </el-form-item>
            </el-form>
            <template #footer>
                <el-button @click="showCreateDialog = false">Cancel</el-button>
                <el-button type="primary" @click="doCreate" :loading="loading">Submit</el-button>
            </template>
        </el-dialog>

        <!-- Update Dialog -->
        <el-dialog
            v-model="showUpdateDialog"
            :title="`Update Article (ID: ${updateForm.id})`"
            width="420"
            :close-on-click-modal="false"
        >
            <el-form ref="updateFormRef" :model="updateForm" :rules="updateRules" label-width="90px">
                <el-form-item label="Title" prop="title">
                    <el-input v-model="updateForm.title" placeholder="Title" />
                </el-form-item>
                <el-form-item label="Content" prop="content">
                    <el-input v-model="updateForm.content" type="textarea" placeholder="Content" :rows="3" />
                </el-form-item>
                <el-form-item label="Student ID" prop="studentId">
                    <el-input v-model="updateForm.studentId" placeholder="Student ID (optional)" />
                </el-form-item>
            </el-form>
            <template #footer>
                <el-button @click="showUpdateDialog = false">Cancel</el-button>
                <el-button type="primary" @click="doUpdate" :loading="loading">Submit</el-button>
            </template>
        </el-dialog>

        <!-- Logs -->
        <details class="log-section">
            <summary>Logs ({{ logs.length }})</summary>
            <div class="log-output" ref="logBox">
                <div v-for="(line, i) in logs" :key="i" class="log-line">{{ line }}</div>
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
    background: linear-gradient(135deg, #43b883 0%, #3594d1 100%);
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
