import { GrpcWebFetchTransport } from '@protobuf-ts/grpcweb-transport'

// demo1kratos backend (StudentService)
export const demo1Transport = new GrpcWebFetchTransport({
    baseUrl: '/demo1kratos-base',
    meta: {
        Authorization: 'TOKEN-888',
    },
})

// demo2kratos backend (ArticleService)
export const demo2Transport = new GrpcWebFetchTransport({
    baseUrl: '/demo2kratos-base',
    meta: {
        Authorization: 'TOKEN-888',
    },
})
