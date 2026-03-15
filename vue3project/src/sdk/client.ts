import { demo1Transport, demo2Transport } from '../api/transport'
import { Sdk学生管理 } from './sdk学生管理'
import { Sdk文章管理 } from './sdk文章管理'

export const sdk学生管理 = new Sdk学生管理(demo1Transport)
export const sdk文章管理 = new Sdk文章管理(demo2Transport)
