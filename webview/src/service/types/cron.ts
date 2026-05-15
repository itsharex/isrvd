// ─── 计划任务 ───

export type CronType = 'SHELL' | 'EXEC' | 'BAT' | 'POWERSHELL' | 'DOCKER_TMP' | 'DOCKER_CTR'

export interface CronTypeInfo {
    value: CronType
    label: string
}

export type CronRuntimeStatus = 'scheduled' | 'disabled' | 'unregistered'

export interface CronJob {
    id: string
    name: string
    schedule: string
    type: CronType
    content: string
    workDir: string
    image?: string
    container?: string
    volumes?: string
    timeout: number
    enabled: boolean
    registered: boolean
    entryId?: number
    runtimeStatus: CronRuntimeStatus
    description: string
    nextRun?: string
    lastRun?: string
}

export interface CronJobCreate {
    name: string
    schedule: string
    type: CronType
    content: string
    workDir?: string
    image?: string
    container?: string
    volumes?: string
    timeout?: number
    enabled: boolean
    description?: string
}

export interface CronJobLog {
    runId: string
    jobId: string
    jobName: string
    startTime: string
    endTime: string
    duration: number
    success: boolean
    output: string
    error?: string
}

export interface CronJobList {
    jobs: CronJob[]
}

export interface CronJobLogList {
    logs: CronJobLog[]
}
