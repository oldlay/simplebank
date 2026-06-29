import type { Account } from './account'
import type { Entry } from './entry'

/** Matches pb.Transfer from the gRPC-Gateway JSON response. */
export interface Transfer {
  id: number
  from_account_id: number
  to_account_id: number
  amount: string | { value: string } // google.type.Decimal
  created_at: string // google.protobuf.Timestamp → RFC 3339
}

/** Matches pb.CreateTransferResponse. */
export interface CreateTransferResponse {
  transfer: Transfer
  from_account: Account
  to_account: Account
  from_entry: Entry
  to_entry: Entry
}
