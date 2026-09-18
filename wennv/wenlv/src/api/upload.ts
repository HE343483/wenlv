/** 上传接口封装 — 获取 OSS 直传签名，并按浏览器 POST 直传协议上传 */
import { post } from './request'

export interface OssPolicy {
  access_key_id: string
  policy: string
  signature: string
  key: string
  host: string
}

export interface UploadPolicyResult {
  policy: OssPolicy
  object_url: string
  expires_in: number
}

/** 获取 OSS 直传签名 */
export function getUploadPolicy(file_name?: string): Promise<UploadPolicyResult> {
  return post<UploadPolicyResult>('/upload/policy', { file_name })
}

/** 将本地文件按 policy 直传至 OSS，返回可访问的 object_url */
export async function uploadToOss(file: File, fileName?: string): Promise<string> {
  const { policy, object_url } = await getUploadPolicy(fileName || file.name)
  const form = new FormData()
  form.append('key', policy.key)
  form.append('policy', policy.policy)
  form.append('OSSAccessKeyId', policy.access_key_id)
  form.append('signature', policy.signature)
  form.append('file', file)

  const resp = await fetch(policy.host, {
    method: 'POST',
    body: form,
  })
  if (!resp.ok) {
    throw new Error('上传失败，请重试')
  }
  return object_url
}