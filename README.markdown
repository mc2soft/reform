MAJOR TODO: add tests for absent pk, multiple-column pk,
and (the killer) multiple-column pk with AUTO_INCREMENT in MySQL (LAST_INSERT_ID() LOL).

Remove reform.xml, generate user_reform.go from user.go? user.go defines struct.

Alternative names: ormogen (ORM generator), gomogen (Go Model Generator)

Problems with mapping
=====================

Go has:
– values
– pointers to values
– convenient notion of "zero value"

Database has:
– values
– NULL values
– nullable columns
– non-nullable columns
– default values for INSERT
– notion of "absence of column name in query" (which is different from both NULLs and default
values)

```go
type User struct {
    Id    int
    Name  string
    Email *string
}
```

```sql
CREATE TABLE users (
    id    serial  PRIMARY KEY,
    name  varchar NOT NULL,
    email varchar
);
```


SELECT
------

It's not possible to scan NULL into non-pointer field.


INSERT
------

1. Insert values and non-nil pointers as is, nil pointers as NULL.
    INSERT INTO users (id, name, email) VALUES (1, 'Alek', NULL); – ok
    INSERT INTO users (id, name, email) VALUES (NULL, 'Alek', NULL); – doesn't work
    In other words, primary key column name must be absent.
    Also this breaks default values other then primary keys, but I can bear it.

2. Insert values and non-nil pointers as is, nil pointers as NULL.
   Special handling for primary keys. Don't care about other default values.
    But db.Insert(&User{}) will insert row with empty ("") Name without error, that's bad.

3. Insert values and non-nil pointers as is, nil pointers as NULL.
   Special handling for primary keys. Don't care about other default values.
   Do not insert zero values.
    But what if one really wants to insert 0 or empty string?

4. Insert values and non-nil pointers as is, nil pointers as NULL.
   Don't care about default values.
   Do not insert zero values which are tagged with "omitempty". Most values should be tagged.
   Primary key is always tagged. Pointers should not be tagged (for UPDATE to work, see below).
    Seems to be ok for INSERT, current implementation.

5. Use pointers for all struct fields, never insert nil pointer.
    u.Name = pointer.ToString("Alek") – weird for all fields


UPDATE
------

1. All columns are present.
    But db.Update(&User{Id: 1, Name: "Alek"}) will set Email to NULL, that's bad.

2. Column is present only if value is present (not zero).
    But then how to change value in nullable column from something to NULL?

3. Column is present only if value is present (not zero) or there is no "omitempty" tag.
   Pointers are not tagged to allow updating to NULL.
    But db.Update(&User{Id: 1, Name: "Alek"}) will set Email to NULL, that's AWFUL.
    But that's the current implementation.
