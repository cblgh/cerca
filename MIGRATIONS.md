# Migrations

This documents migrations for breaking database changes. These are intended to be as few as
possible, but sometimes they are necessary.

To run migrations, you call `cerca migrate`. You can access help for the migrate tool by running

```
go build ./cmd/cerca
./cerca migrate --help
```

Get a full list of migrations by passing `--list`:

```
./cerca migrate --list
```

## [2024-07-20] Private threads

Add a column to `database.Thread` to signal whether or not the thread is private.

For more details, see [database/migrations.go](./database/migrations.go).

Build cerca then run `cerca migrate` accordingly:

```
go build ./cmd/cerca
./cerca migrate --database path-to-your-forum.db --migration 2024-02-thread-private-migration
```

## [2024-01-16] Migrating password hash libraries

To support 32 bit architectures, such as running Cerca on an older Raspberry Pi, the password
hashing library used previously
[github.com/synacor/argon2id](https://github.com/synacor/argon2id) was swapped out for
[github.com/matthewhartstonge/argon2](https://github.com/matthewhartstonge/argon2).

The password hashing algorithm remains the same, so users **do not** need to reset their
password, but the old encoding format needed migrating in the database. 

The new library also has a stronger default time parameter, which will be used for new
passwords while old passwords will use the time parameter stored together with their
database record.

For more details, see [database/migrations.go](./database/migrations.go).

Build cerca then run `cerca migrate` accordingly:

```
go build ./cmd/cerca
./cerca migrate --database path-to-your-forum.db --migration 2024-01-password-hash-migration   
```


