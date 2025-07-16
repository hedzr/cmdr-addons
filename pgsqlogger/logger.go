package psqlogger

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"sync/atomic"

	logz "github.com/hedzr/logg/slog"

	_ "github.com/lib/pq"
)

var _ logz.LogWriter = (*PGSQLLogger)(nil)

// New returns a pgsql logger as a logg/slog's [logz.LogWriter].
//
// How to use it:
//
//	func cmdrCommandActionHandler(ctx context.Context, cmd cli.Cmd, args []string) (err error) {
//	    conf := cmdr.Set().WithPrefix("resources.db.postgres")
//	    host := conf.MustString("host", "127.0.0.1")
//	    port := conf.MustInt("port", 5432)
//	    user := conf.MustString("user", "postgres")
//	    password := conf.MustString("password", "postgres")
//	    dbname := conf.MustString("db", "postgres")
//
//	    logger := psqlogger.New()
//	    err = logger.Open(ctx, &psqlogger.ConnectOpt{
//	        Host:host,Port:port,User:user,Password:password,DBName:dbname,
//	    })
//
//	    logz.Default().AddWriter(logger)
//	    logz.Default().AddErrorWriter(logger)
//
//	    // the logger will be closed automatically
//	    basics.RegisterClosers(logger)
//
//	    logz.InfoContext(ctx, "A sample logging line here", "cmd", cmd)
//	    return
//	}
//
// To enable the above sample code, make sure `db_logging` table has been created by:
//
//	create table db_logging (
//	    log_id SERIAL PRIMARY KEY,
//	    tm timestamp,
//	    msg jsonb -- text -- varchar(120)
//	);
func New() *PGSQLLogger {
	return &PGSQLLogger{sqlog: &sqlog{}}
}

type ConnectOpt struct {
	Host     string `json:"host,omitempty"`
	Port     int    `json:"port,omitempty"`
	User     string `json:"user,omitempty"`
	Password string `json:"password,omitempty"`
	DBName   string `json:"db,omitempty"`
}

type PGSQLLogger struct {
	*sqlog
	stmt        *sql.Stmt
	closed      int32
	cachedLevel logz.Level
}

func (s *PGSQLLogger) Open(ctx context.Context, opt *ConnectOpt) (err error) {
	if err = s.Connect(ctx, opt); err != nil {
		return
	}
	err = s.Prepare(ctx)
	return
}

func (s *PGSQLLogger) Close() (err error) {
	if atomic.CompareAndSwapInt32(&s.closed, 0, 1) {
		logz.Verbose("[PGSQLLogger] closing...")
		if st := s.stmt; st != nil {
			s.stmt = nil
			err = st.Close()
		}
		if s.sqlog != nil {
			s.sqlog.Close()
			if s.err != nil {
				err = s.err
			}
		}
		logz.Verbose("[PGSQLLogger] closed...")
	}
	return
}

func (s *PGSQLLogger) Prepare(ctx context.Context) (err error) {
	// query := `PREPARE insert_logging_line_plan (text) AS
	// 	INSERT INTO db_logging (tm, msg) VALUES (CURRENT_TIMESTAMP, $1);`
	// query := `INSERT INTO db_logging (tm, level, msg) VALUES (CURRENT_TIMESTAMP, $1, $2);`
	query := `INSERT INTO db_logging (tm, msg) VALUES (CURRENT_TIMESTAMP, $1);`
	s.stmt, err = s.PrepareContext(ctx, query)
	return
}

func (s *PGSQLLogger) Write(data []byte) (n int, err error) {
	if s.stmt != nil {
		var res sql.Result
		if res, err = s.stmt.Exec(string(data)); err != nil {
			slog.Error("write log line into pgsql db failed", "log", string(data), "sql-result", res, "err", err)
		} else {
			n = len(data)
		}
	}
	return
}

func (s *PGSQLLogger) SetLevel(level logz.Level) {
	s.cachedLevel = level
}

type sqlog struct {
	conn *sql.DB
	err  error
}

func (s *sqlog) Close() {
	// here's cleanup operations to free the conn object
	if s.conn != nil {
		if s.err = s.conn.Close(); s.err == nil {
			s.conn = nil
			logz.Verbose(`database connection closed`)
		}
	}
}

func (s *sqlog) Open(ctx context.Context, opt *ConnectOpt) (err error) {
	return s.Connect(ctx, opt)
}

func (s *sqlog) Connect(ctx context.Context, opt *ConnectOpt) (err error) {
	// do stuffs to open connection to database here

	// // locate to `app.resources.db.postgres`
	// conf := cmdr.Set().WithPrefix("resources.db.postgres")
	// host := conf.MustString("host", "127.0.0.1")
	// port := conf.MustInt("port", 5432)
	// user := conf.MustString("user", "postgres")
	// password := conf.MustString("password", "postgres")
	// dbname := conf.MustString("db", "postgres")

	dsn := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		opt.Host, opt.Port, opt.User, opt.Password, opt.DBName)

	logz.Verbose(`opening database...`, "DSN", dsn)

	s.conn, err = sql.Open("postgres", dsn)
	if err != nil {
		return
	}

	logz.Verbose(`database connection opened`)
	err = s.conn.Ping()

	_ = ctx
	return
}

// QueryContext executes a query that returns rows, typically a SELECT.
// The args are for any placeholder parameters in the query.
func (s *sqlog) QueryContext(ctx context.Context, query string, args ...any) (rows *sql.Rows, err error) {
	if s.conn == nil {
		return nil, fmt.Errorf("database connection not ready")
	}
	return s.conn.QueryContext(ctx, query, args...)
}

// ExecContext executes a query without returning any rows.
// The args are for any placeholder parameters in the query.
func (s *sqlog) ExecContext(ctx context.Context, query string, args ...any) (res sql.Result, err error) {
	if s.conn == nil {
		return nil, fmt.Errorf("database connection not ready")
	}
	res, err = s.conn.ExecContext(ctx, query, args)
	return
}

// PrepareContext creates a prepared statement for later queries or executions.
// Multiple queries or executions may be run concurrently from the
// returned statement.
// The caller must call the statement's [*sql.Stmt.Close] method
// when the statement is no longer needed.
//
// The provided context is used for the preparation of the statement, not for the
// execution of the statement.
func (s *sqlog) PrepareContext(ctx context.Context, query string) (stmt *sql.Stmt, err error) {
	stmt, err = s.conn.PrepareContext(ctx, query)
	return
}

// BeginTx starts a transaction.
//
// The provided context is used until the transaction is committed or rolled back.
// If the context is canceled, the sql package will roll back
// the transaction. [sql.Tx.Commit] will return an error if the context provided to
// BeginTx is canceled.
//
// The provided [sql.TxOptions] is optional and may be nil if defaults should be used.
// If a non-default isolation level is used that the driver doesn't support,
// an error will be returned.
func (s *sqlog) BeginTx(ctx context.Context, opts *sql.TxOptions) (tx *sql.Tx, err error) {
	tx, err = s.conn.BeginTx(ctx, opts)
	return
}
