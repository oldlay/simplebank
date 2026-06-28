/** Matches pb.Entry from the gRPC-Gateway JSON response. */
export interface Entry {
  id: number
  account_id: number
  amount: string | { value: string } // google.type.Decimal
}
