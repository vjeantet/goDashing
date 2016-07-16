# BytePool

A pool for `[]byte`.

## Usage:

The pool is thread-safe.

```go
// create a pool of 16 items, each item has a size of 1024
var pool = bytepool.New(1024, 16)


bytes := pool.Checkout()
defer bytes.Release()
bytes.Write([]byte("hello"))
fmt.Prinltn(bytes.String())
```

## Pool Growth
Getting an item from the pool is non-blocking. If the pool is depleted, a new item will be created. However, such dynamically created items are not added back to the pool on release. In other words, the # of items within the pool is fixed.

You can call the `Depleted()` method for count of how often the pool was depleted. If it's frequently greater than 0, consider increasing the count of items your pool holds.

## Item Growth
Items are created with an initial size. Adding more data than this size will cause the item to internally and efficiently convert itself to a `bytes.Buffer`. However, the growth is not permanent: on `Release` the initial allocation is re-established (without needing to do a new allocation).

In other words, the total memory used by the bytepool should be size * count. For our above example, that's 16KB. The size will increase as needed, but will always revert back to 16KB (and that 16KB is only initiated once, on startup).

You can call the `Expanded()` method for a count of how often items were forced to grow beyond the initial size. If it's frequently greater than 0, consider increasing the initial size of the items.


## Pool Methods:
* `New(size, count)` - Creates a new pool of `count` items, each initialize to hold `size` bytes (but able to grow as needed)
* `Checkout() *Bytes` - Gets a item which you can write/read from
* `Depleted() int64` - How often the pool was empty. Calling this resets the counter
* `Expanded() int64` - How often items were forced to grow beyond their initial size. Calling this resets the counter
* `Stats() map[string]int64` - `Depleted` and `Expanded` in a map

## Item Methods:
* `Write(data []byte) (n int, err error)`
* `WriteByte(data byte) error`
* `WriteString(data string) (n int, err error)`
* `Bytes() []byte`
* `String() string`
* `Len() int`
* `ReadFrom(r io.Reader) (n int64, err error)`
* `ReadNFrom(n int64, r io.Reader) (m int64, err error)`
* `Read(data []byte) (int, error)`
* `Release()` or `Close()` - Resets and releases the item back to the pool.
* `Reset()` - Resets the item without releasing it back to the pool
* `Position(n uint)` - Moves to the specified absolute position. This will grow the buffer if needed.

## Numeric Encoding
The `WriteUint16`, `WriteUint32` and `WriteUint64` methods can be used to write integers
in big endian.

To write using little endian, create a pool using `NewEndian` or an individual item using `NewEndianBytes` and pass the `binary.LittleEndian` object from the stdlib "encoding/binary" package.

Corresponding `ReadUint16`, `ReadUint32` and `ReadUint64` are availabl. They return `(n, error)` where error will be `io.EOF` if not enough data is available.

# Each
It's possible to pre-fill byte items within the pool through the use of the pool's `Each` and the item's `Position` functions. You have to take special care to properly `Position` the item on each checkout.

If you pre-fill the item with more data than the initial capacity, the data will be lost. It makes no sense that you'd define a pool with items having a capacity of 50 bytes., but pre-fill 100 bytes.

For example, say we were writing into a buffer where bytes 4-8 were always the value 30, we could do:

```go
//setup
pool := bytepool.New(256, 10)
pool.Each(func(b *bytepool.Bytes) {
  b.Position(4)
  b.WriteInt32(30)
})


// code that uses the buffer
bytes := pool.Checkout()
bytes.WriteInt32(size) //write into bytes[0:4]
bytes.Position(8)    //skip bytes[4:8] which we already filled
//continue
```
