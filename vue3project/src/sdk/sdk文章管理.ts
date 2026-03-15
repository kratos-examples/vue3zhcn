import type { RpcTransport } from '@protobuf-ts/runtime-rpc'

import { ArticleServiceClient } from '../rpc/demo2/article/article.client'
import {
    CreateArticleRequest,
    UpdateArticleRequest,
    DeleteArticleRequest,
    GetArticleRequest,
    ListArticlesRequest,
} from '../rpc/demo2/article/article'

// ========================== 请求类型定义 ==========================

export interface Req创建文章 {
    v标题: string
    v内容: string
    v学生编号: string
}

export interface Req更新文章 {
    v编号: string
    v标题: string
    v内容: string
    v学生编号: string
}

export interface Req删除文章 {
    v编号: string
}

export interface Req获取文章 {
    v编号: string
}

export interface Req文章列表 {
    v页码: number
    v每页数量: number
}

// ========================== 响应类型定义 ==========================

export interface T文章信息 {
    s编号: string
    s标题: string
    s内容: string
    s学生编号: string
}

export interface Res创建文章 {
    s文章: T文章信息 | undefined
}

export interface Res更新文章 {
    s文章: T文章信息 | undefined
}

export interface Res删除成功 {
    b成功: boolean
}

export interface Res获取文章 {
    s文章: T文章信息 | undefined
}

export interface Res文章列表 {
    s文章列表: T文章信息[]
    s总数: number
}

// ========================== SDK 类定义 ==========================

function cnv文章信息(a: { id: string; title: string; content: string; studentId: string } | undefined): T文章信息 | undefined {
    if (!a) return undefined
    return {
        s编号: a.id,
        s标题: a.title,
        s内容: a.content,
        s学生编号: a.studentId,
    }
}

export class Sdk文章管理 {
    private rpc客户端: ArticleServiceClient

    constructor(transport: RpcTransport) {
        this.rpc客户端 = new ArticleServiceClient(transport)
    }

    async act创建文章(req: Req创建文章): Promise<Res创建文章> {
        const res = await this.rpc客户端.createArticle(
            CreateArticleRequest.create({
                title: req.v标题,
                content: req.v内容,
                studentId: req.v学生编号,
            }),
            {},
        )
        return { s文章: cnv文章信息(res.data.article) }
    }

    async act更新文章(req: Req更新文章): Promise<Res更新文章> {
        const res = await this.rpc客户端.updateArticle(
            UpdateArticleRequest.create({
                id: req.v编号,
                title: req.v标题,
                content: req.v内容,
                studentId: req.v学生编号,
            }),
            {},
        )
        return { s文章: cnv文章信息(res.data.article) }
    }

    async act删除文章(req: Req删除文章): Promise<Res删除成功> {
        const res = await this.rpc客户端.deleteArticle(
            DeleteArticleRequest.create({ id: req.v编号 }),
            {},
        )
        return { b成功: res.data.success }
    }

    async act获取文章(req: Req获取文章): Promise<Res获取文章> {
        const res = await this.rpc客户端.getArticle(
            GetArticleRequest.create({ id: req.v编号 }),
            {},
        )
        return { s文章: cnv文章信息(res.data.article) }
    }

    async act文章列表(req: Req文章列表): Promise<Res文章列表> {
        const res = await this.rpc客户端.listArticles(
            ListArticlesRequest.create({
                page: req.v页码,
                pageSize: req.v每页数量,
            }),
            {},
        )
        return {
            s文章列表: res.data.articles.map((a) => cnv文章信息(a)!),
            s总数: res.data.count,
        }
    }
}
