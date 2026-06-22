export interface User {
  id: string
  name: string
  email: string
  role: string
}

export interface Asset {
  id: string
  domain: string
  port: number
  label: string
  status: string
  last_scanned_at: string | null
  current_risk_score: number
  current_pqc_score: number
  created_at: string
  scans?: Scan[]
}

export interface Certificate {
  id: string
  scan_id: string
  asset_id: string
  asset?: Asset
  subject: string
  common_name: string
  issuer: string
  serial_number: string
  not_before: string
  not_after: string
  public_key_algorithm: string
  signature_algorithm: string
  key_size: number
  chain_valid: boolean
  hostname_valid: boolean
  self_signed: boolean
  subject_alt_names: string[]
  fingerprint_sha256: string
  pem?: string
  created_at: string
}

export interface Finding {
  id: string
  scan_id: string
  asset_id: string
  asset?: Asset
  type: string
  severity: 'critical' | 'high' | 'medium' | 'low' | 'info'
  title: string
  description: string
  remediation: string
  evidence: string
  status: string
  created_at: string
}

export interface PQCAssessment {
  id: string
  scan_id: string
  asset_id: string
  asset?: Asset
  score: number
  grade: string
  quantum_vulnerable: boolean
  rsa_dependency: boolean
  ecc_dependency: boolean
  tls13_enabled: boolean
  legacy_tls_supported: boolean
  certificate_rotation_ready: boolean
  rationale: string[]
  created_at: string
}

export interface Scan {
  id: string
  asset_id: string
  asset?: Asset
  status: string
  started_at: string | null
  completed_at: string | null
  duration_ms: number
  risk_score: number
  pqc_score: number
  tls_version: string
  cipher_suite: string
  error_message?: string
  certificate?: Certificate
  findings?: Finding[]
  pqc_assessment?: PQCAssessment
  created_at: string
}

export interface DashboardData {
  summary: {
    asset_count: number
    assessed_count: number
    average_risk_score: number
    average_pqc_score: number
    critical_findings: number
  }
  assets: Asset[]
  recent_scans: Scan[]
  priority_findings: Finding[]
}

export interface ReportSummary {
  generated_at: string
  findings_by_severity: Array<{ severity: string; count: number }>
  certificates_by_algorithm: Array<{ algorithm: string; count: number }>
  certificates_expiring_90_days: number
}

