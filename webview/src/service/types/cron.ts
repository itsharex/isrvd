// ─── 计划任务 ───

export type CronType = 'SHELL' | 'EXEC' | 'BAT' | 'POWERSHELL'

export interface CronTypeInfo {
    value: CronType
    label: string
}

export interface CronJob {
    id: string
    name: string
    schedule: string
    type: CronType
    content: string
    workDir: string
    timeout: number
    enabled: boolean
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
    timeout?: number
    enabled: boolean
    description?: string
}

export interface CronJobLog {
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
