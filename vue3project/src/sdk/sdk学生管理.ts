import type { RpcTransport } from '@protobuf-ts/runtime-rpc'
import type { GrpcToHttpWebSocket } from '@yylego/grpc-to-http'

import { StudentServiceClient } from '../rpc/demo1/student/student.client'
import { CreateStudentRequest, UpdateStudentRequest, DeleteStudentRequest, GetStudentRequest, ListStudentsRequest } from '../rpc/demo1/student/student'
import type { GetStudentReply } from '../rpc/demo1/student/student'

// ========================== 请求类型定义 ==========================

export interface Req创建学生 {
    v名字: string
    v年龄: number
    v班级: string
}

export interface Req更新学生 {
    v编号: string
    v名字: string
    v年龄: number
    v班级: string
}

export interface Req删除学生 {
    v编号: string
}

export interface Req获取学生 {
    v编号: string
}

export interface Req学生列表 {
    v页码: number
    v每页数量: number
}

// ========================== 响应类型定义 ==========================

export interface T学生信息 {
    s编号: string
    s名字: string
    s年龄: number
    s班级: string
}

export interface Res创建学生 {
    s学生: T学生信息 | undefined
}

export interface Res更新学生 {
    s学生: T学生信息 | undefined
}

export interface Res删除成功 {
    b成功: boolean
}

export interface Res获取学生 {
    s学生: T学生信息 | undefined
}

export interface Res学生列表 {
    s学生列表: T学生信息[]
    s总数: number
}

// ========================== 实时查询连接 ==========================

// Conn学生查询 包着双向流式的 LookupStudents：一条 WebSocket 连接上每发一个编号，就回来那个学生，用 for await 逐条读
export class Conn学生查询 implements AsyncIterable<Res获取学生> {
    private rpc连接: GrpcToHttpWebSocket<GetStudentRequest, GetStudentReply>

    constructor(rpc连接: GrpcToHttpWebSocket<GetStudentRequest, GetStudentReply>) {
        this.rpc连接 = rpc连接
    }

    act发送编号(req: Req获取学生): void {
        this.rpc连接.send(GetStudentRequest.create({ id: req.v编号 }))
    }

    // 告诉服务端不再发编号，服务端返回后正常关闭连接
    act结束发送(): void {
        this.rpc连接.closeSend()
    }

    act关闭连接(): void {
        this.rpc连接.close()
    }

    async *[Symbol.asyncIterator](): AsyncGenerator<Res获取学生, void, undefined> {
        for await (const reply of this.rpc连接) {
            yield { s学生: cnv学生信息(reply.student) }
        }
    }
}

// ========================== SDK 类定义 ==========================

function cnv学生信息(s: { id: string; name: string; age: number; className: string } | undefined): T学生信息 | undefined {
    if (!s) return undefined
    return {
        s编号: s.id,
        s名字: s.name,
        s年龄: s.age,
        s班级: s.className,
    }
}

export class Sdk学生管理 {
    private rpc客户端: StudentServiceClient

    constructor(transport: RpcTransport) {
        this.rpc客户端 = new StudentServiceClient(transport)
    }

    async act创建学生(req: Req创建学生): Promise<Res创建学生> {
        const res = await this.rpc客户端.createStudent(
            CreateStudentRequest.create({
                name: req.v名字,
                age: req.v年龄,
                className: req.v班级,
            }),
            {},
        )
        return { s学生: cnv学生信息(res.data.student) }
    }

    async act更新学生(req: Req更新学生): Promise<Res更新学生> {
        const res = await this.rpc客户端.updateStudent(
            UpdateStudentRequest.create({
                id: req.v编号,
                name: req.v名字,
                age: req.v年龄,
                className: req.v班级,
            }),
            {},
        )
        return { s学生: cnv学生信息(res.data.student) }
    }

    async act删除学生(req: Req删除学生): Promise<Res删除成功> {
        const res = await this.rpc客户端.deleteStudent(DeleteStudentRequest.create({ id: req.v编号 }), {})
        return { b成功: res.data.success }
    }

    async act获取学生(req: Req获取学生): Promise<Res获取学生> {
        const res = await this.rpc客户端.getStudent(GetStudentRequest.create({ id: req.v编号 }), {})
        return { s学生: cnv学生信息(res.data.student) }
    }

    async act学生列表(req: Req学生列表): Promise<Res学生列表> {
        const res = await this.rpc客户端.listStudents(
            ListStudentsRequest.create({
                page: req.v页码,
                pageSize: req.v每页数量,
            }),
            {},
        )
        return {
            s学生列表: res.data.students.map((s) => cnv学生信息(s)!),
            s总数: res.data.count,
        }
    }

    act连接查询(): Conn学生查询 {
        return new Conn学生查询(this.rpc客户端.lookupStudents())
    }
}
