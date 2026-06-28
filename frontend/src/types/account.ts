/** Matches pb.Account from the gRPC-Gateway JSON response. */
export interface Account {
  id: number
  owner: string
  currency: string
  amount: string | { value: string } // google.type.Decimal
  create_at: string // google.protobuf.Timestamp → RFC 3339
}

/** Extract the decimal amount as a plain string, e.g. "1500.00" */
export function getAmount(account: Account): string {
  if (typeof account.amount === 'object' && account.amount !== null) {
    return account.amount.value
  }
  return account.amount
}

/** Wrap a decimal value for the gRPC request body. */
export function toDecimalValue(raw: string): { value: string } {
  return { value: raw }
}

export interface CreateAccountResponse { account: Account }
export interface UpdateAccountResponse { account: Account }
export interface DeleteAccountResponse { delete_result: string }
export interface ListAccountResponse { accounts: Account[] }
