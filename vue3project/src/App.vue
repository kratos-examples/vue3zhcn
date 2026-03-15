<script setup lang="ts">
import { ref } from 'vue'
import StudentDemo from './components/StudentDemo.vue'
import ArticleDemo from './components/ArticleDemo.vue'

const activeTab = ref('student')
</script>

<template>
    <div class="app">
        <div class="app-header">
            <h1>Vue3 + Kratos Demo</h1>
            <p class="subtitle">gRPC-to-HTTP integration via protobuf-ts</p>
        </div>

        <el-card class="main-card" shadow="always">
            <el-tabs v-model="activeTab">
                <el-tab-pane label="StudentService" name="student">
                    <StudentDemo />
                </el-tab-pane>
                <el-tab-pane label="ArticleService" name="article">
                    <ArticleDemo />
                </el-tab-pane>
            </el-tabs>
        </el-card>
    </div>
</template>

<style scoped>
.app {
    max-width: 960px;
    margin: 0 auto;
    padding: 32px 24px;
}
.app-header {
    text-align: center;
    margin-bottom: 20px;
}
h1 {
    margin: 0 0 4px 0;
    color: #303133;
}
.subtitle {
    margin: 0;
    color: #909399;
    font-size: 14px;
}
.main-card {
    border-radius: 12px;
}
</style>
