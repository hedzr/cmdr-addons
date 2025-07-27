# pgsqlogger

This plugin for [`hedzr/logg/slog`](https://github.com/hedzr/logg) adds supports to write logging lines into database.

## Usages

The codes is (based cmdr.v2):

```go
func (s *convertCmd) initDBLogger(ctx context.Context, cmd cli.Cmd) (err error) {
	conf := cmd.Set().WithPrefix("resources.db.postgres")
	host := conf.MustString("host", "127.0.0.1")
	port := conf.MustInt("port", 5432)
	user := conf.MustString("user", "postgres")
	password := conf.MustString("password", "postgres")
	dbname := conf.MustString("db-name", "postgres")

	logger := pgsqlogger.New()
	err = logger.Open(ctx, &pgsqlogger.ConnectOpt{
		Host: host, Port: port, User: user, Password: password, DbName: dbname,
	})
	if err != nil {
		return nil // ignore pgsql db logger, don't interrupt main program's running
	}

	// to save log into postgresql db, we need a json format output.
	logz.Default().SetMode(logz.ModeJSON) // set to json format
	logz.Default().AddWriter(logger)      // added as a cloned writer
	logz.Default().AddErrorWriter(logger) // added as a cloned writer

	// the logger will be closed automatically. cmdr will call to basics.Closers().CloseAll.
	basics.RegisterClosers(logger)

	// a sample logging line just for test purpose.
	logz.InfoContext(ctx, "A sample logging line here", "cmd", cmd)
	return
}

// covert will be used in convertCmd.Add(...), which is a standard cmdr.v2 command Action Handler.
func (s *convertCmd) convert(ctx context.Context, cmd cli.Cmd, args []string) (err error) {
	if err = s.initDBLogger(ctx, cmd); err != nil {
		return
	}
    // ...
    return
}
```

Here we load those config items via `hedzr/store`, which has been integrated with `cmdr.v2`. The corresponding TOML file content is:

```toml
[resources.db]
[resources.db.postgres]
    # dsn = ""
    db-name    = "postgres"
    host       = "localhost"
    password   = "postgres"
    port       = 5432
    table-name = "" # if you wanna customize it
    user       = "postgres"
```

You can also use loading into struct feature in [`hedzr/Store`](https://github.com/hedzr/store):

```go
	var opt pgsqlogger.ConnectOpt
	if err = cmd.Set().To("resources.db.postgres", &opt); err != nil {
		return
	}

	logger := pgsqlogger.New()
	err = logger.Open(ctx, &opt)
	if err != nil {
		return nil // ignore pgsql db logger, don't interrupt main program's running
	}
```

The table `db_logging` must be present in the postgres database service:

```sql
drop table db_logging;
create table db_logging (
    log_id SERIAL PRIMARY KEY,
    tm timestamp,
    msg jsonb -- text -- varchar(120)
);
```

Note the messages are stored as a `jsonb` field so we can filter the logging lines with json field. For example, the following query looks up all error lines:

```sql
select msg->'level', msg->'msg', msg as full_json_data from db_logging where msg->'level' = '"debug"';
```

### Customize

~~TODO the further customizable options will be added in recent versions~~.

#### WithTableName

To change the logging table name is possible now. By using `WithTableName(tableName)` as a option of `New(opts...)`, you can customize it.
