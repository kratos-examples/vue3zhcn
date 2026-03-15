<script setup lang="ts">
import { ref, reactive, nextTick, onMounted } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'
import { demo1Transport } from '../api/transport'
import { StudentServiceClient } from '../rpc/demo1/student/student.client'
import {
    CreateStudentRequest,
    UpdateStudentRequest,
    DeleteStudentRequest,
    GetStudentRequest,
    ListStudentsRequest,
} from '../rpc/demo1/student/student'
import type { StudentInfo } from '../rpc/demo1/student/student'
import { showErrorDialog } from '../utils/error'
import { showSuccess } from '../utils/message'

const client = new StudentServiceClient(demo1Transport)

const students = ref<StudentInfo[]>([])
const logs = ref<string[]>([])
const logBox = ref<HTMLElement | null>(null)
const loading = ref(false)

// create dialog
const showCreateDialog = ref(false)
const createFormRef = ref<FormInstance>()
const createForm = reactive({ name: '', age: 18, className: '' })
const createRules = reactive<FormRules>({
    name: [{ required: true, message: 'Name is required', trigger: 'blur' }],
})

// update dialog
const showUpdateDialog = ref(false)
const updateFormRef = ref<FormInstance>()
const updateForm = reactive({ id: '', name: '', age: 0, className: '' })
const updateRules = reactive<FormRules>({
    name: [{ required: true, message: 'Name is required', trigger: 'blur' }],
})

function log(msg: string) {
    logs.value.push(`[${new Date().toLocaleTimeString()}] ${msg}`)
    if (logs.value.length > 50) logs.value.shift()
    nextTick(() => {
        if (logBox.value) logBox.value.scrollTop = logBox.value.scrollHeight
    })
}

function openCreate() {
    createForm.name = ''
    createForm.age = 18
    createForm.className = ''
    showCreateDialog.value = true
    nextTick(() => createFormRef.value?.clearValidate())
}

async function doCreate() {
    const valid = await createFormRef.value?.validate().catch(() => false)
    if (!valid) return
    loading.value = true
    try {
        const request = CreateStudentRequest.create({
            name: createForm.name.trim(),
            age: createForm.age,
            className: createForm.className.trim(),
        })
        const response = await client.createStudent(request, {})
        const s = response.data.student
        showSuccess(`Created: id=${s?.id}, name=${s?.name}`)
        log(`Created: id=${s?.id}, name=${s?.name}`)
        await doList()
        showCreateDialog.value = false
    } catch (caught: unknown) {
        showErrorDialog(caught)
        log(`Create FAIL: ${caught}`)
    }
    loading.value = false
}

function openUpdate(s: StudentInfo) {
    updateForm.id = s.id
    updateForm.name = s.name
    updateForm.age = s.age
    updateForm.className = s.className
    showUpdateDialog.value = true
    nextTick(() => updateFormRef.value?.clearValidate())
}

async function doUpdate() {
    const valid = await updateFormRef.value?.validate().catch(() => false)
    if (!valid) return
    loading.value = true
    try {
        const request = UpdateStudentRequest.create({
            id: updateForm.id,
            name: updateForm.name.trim(),
            age: updateForm.age,
            className: updateForm.className.trim(),
        })
        const response = await client.updateStudent(request, {})
        const s = response.data.student
        showSuccess(`Updated: id=${s?.id}, name=${s?.name}`)
        log(`Updated: id=${s?.id}, name=${s?.name}`)
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
        const request = ListStudentsRequest.create({})
        const response = await client.listStudents(request, {})
        students.value = response.data.students
        log(`Loaded ${response.data.students.length} students`)
    } catch (caught: unknown) {
        showErrorDialog(caught)
        log(`List FAIL: ${caught}`)
    }
    loading.value = false
}

async function doGet(id: string) {
    loading.value = true
    try {
        const request = GetStudentRequest.create({ id })
        const response = await client.getStudent(request, {})
        const s = response.data.student
        log(`Get: id=${s?.id}, name=${s?.name}, age=${s?.age}, class=${s?.className}`)
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
        const request = DeleteStudentRequest.create({ id })
        await client.deleteStudent(request, {})
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
            <h2>StudentService</h2>
            <span class="port">(demo1kratos :8001)</span>
        </div>

        <!-- Student List -->
        <div class="list-card">
            <div class="list-header">
                <h3>Students</h3>
                <div class="header-actions">
                    <el-button type="primary" @click="openCreate" :loading="loading" size="small">Create</el-button>
                    <el-button @click="doList" :loading="loading" size="small">Refresh</el-button>
                </div>
            </div>
            <el-table :data="students" v-loading="loading" stripe border v-if="students.length > 0">
                <el-table-column prop="id" label="ID" width="80" align="center" />
                <el-table-column prop="name" label="Name" align="center" />
                <el-table-column prop="age" label="Age" width="80" align="center" />
                <el-table-column prop="className" label="Class" align="center" />
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
            <div v-else class="empty-hint">No students yet. Click Create to add one.</div>
        </div>

        <!-- Create Dialog -->
        <el-dialog v-model="showCreateDialog" title="Create Student" width="420" :close-on-click-modal="false">
            <el-form ref="createFormRef" :model="createForm" :rules="createRules" label-width="80px">
                <el-form-item label="Name" prop="name">
                    <el-input v-model="createForm.name" placeholder="Name" />
                </el-form-item>
                <el-form-item label="Age" prop="age">
                    <el-input-number v-model="createForm.age" :min="1" :max="200" />
                </el-form-item>
                <el-form-item label="Class" prop="className">
                    <el-input v-model="createForm.className" placeholder="Class Name" />
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
            :title="`Update Student (ID: ${updateForm.id})`"
            width="420"
            :close-on-click-modal="false"
        >
            <el-form ref="updateFormRef" :model="updateForm" :rules="updateRules" label-width="80px">
                <el-form-item label="Name" prop="name">
                    <el-input v-model="updateForm.name" placeholder="Name" />
                </el-form-item>
                <el-form-item label="Age" prop="age">
                    <el-input-number v-model="updateForm.age" :min="1" :max="200" />
                </el-form-item>
                <el-form-item label="Class" prop="className">
                    <el-input v-model="updateForm.className" placeholder="Class Name" />
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
