package sqltracing

// SQLOp is the name of an sql operation.
// It can be passed to the WithOpsExcluded() option
type SQLOp string

const (
	OpSQLPrepare    SQLOp = "sql-prepare"
	OpSQLConnExec   SQLOp = "sql-conn-exec"
	OpSQLConnQuery  SQLOp = "sql-conn-query"
	OpSQLStmtExec   SQLOp = "sql-stmt-exec"
	OpSQLStmtQuery  SQLOp = "sql-stmt-query"
	OpSQLStmtClose  SQLOp = "sql-stmt-close"
	OpSQLTxBegin    SQLOp = "sql-tx-begin"
	OpSQLTxCommit   SQLOp = "sql-tx-commit"
	OpSQLTxRollback SQLOp = "sql-tx-rollback"
	OpSQLRowsNext   SQLOp = "sql-rows-next"
	OpSQLRowsClose  SQLOp = "sql-rows-close"
	OpSQLPing       SQLOp = "sql-ping"
	OpSQLConnect    SQLOp = "sql-connect"
)

// String returns the string representation of SQLOp.
func (s SQLOp) String() string {
	return string(s)
}
