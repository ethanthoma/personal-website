# Abusing Use

> tldr; a simple guide on abusing gleam's `use` keyword to make useful,
> non-obvious programming monstrosities.

[Gleam](https://gleam.run/) is a simple language.

Too simple for me.

I love zig comptime and hell, I even do runtime checks on Python type
annotations. I love metaprogramming. It's unbelievable useful, but gleam doesn't
have it. Sure, the author of the lang
[wrote about it](https://lpil.uk/blog/how-to-add-metaprogramming-to-gleam/) and
perhaps, once LLMs have replaced all coders, it will have it. Till then, it is
probably better bet to use a similar lang that does have it (like my beloved
[flix](https://flix.dev/)).

**However**, gleam does have one amazing feature: `use`.

## WTF is `use`

The `use` syntax is a bit of syntactic sugar. What it does is sort-of unwrap
callback functions. Let's taking a common function one would use, `list.map`.

```gleam
pub fn map(list: List(a), with fun: fn(a) -> b) -> List(b) {
  map_loop(list, fun, [])
}
```

The `map` function takes in two parameters, the list it is operating over,
`list` and a mapping function, `fun`. Let's say we had a list of ints we wanted
to increment, we could

```gleam
let ints = [0, 1, 2, 3, 4]

let inc = fn(int: Int) -> Int {
    int + 1
}

echo list.map(ints, inc)
```

which prints out

```
[1, 2, 3, 4, 5]
```

In this example, we created a lambda function (which we named `inc`). Instead,
we can "unwrap" it with `use`:

```gleam
let ints = [0, 1, 2, 3, 4]

echo {
    use int <- list.map(ints, inc)
    int + 1 
}
```

Now what matters is that everything _after_ the `use` is now **inside** the
lambda function. Sort-of like an implicit scope. In otherwords, we _plop_ the
callback scope into the current scope.

## Common Uses

`use` is used in two common ways: `Result`/`Option` types and `bool.guard`.

Gleam uses `Result` for errors and `Option` for `nullables`. Often, one wants to
do something when the function _doesn't_ error. For example:

```gleam
let database: Result(String, Nil) = envoy.get("DB")

result.try(database, fn(database: String) {
    let query: Result(Dynamic, Nil) = sql.query(db, "SELECT * FROM Users;")
    result.map(query, fn(query: Dynamic) {
        let users: List(User) = decode_to_user(decode.list(query))
        echo users
    })
})
```

This is...pretty hard to read. It isn't terrible but as the nesting gets worse
and worse...so does the readability. In fact, this problem has a name:
[callback hell](http://callbackhell.com/).

`use` lets us avoid this. We can rewrite it as:

```gleam
let database: Result(String, Nil) = envoy.get("DB")

use database: String <- result.try(database)

let query: Result(Dynamic, Nil) = sql.query(db, "SELECT * FROM Users;")

use query: Dynamic <- result.map(query)

let users: List(User) = decode_to_user(decode.list(query))

echo users
}
```

This we can call callback heaven.

Another common way for using `use` is for early returns. Lets take this example:

```gleam
let should_return_early = True

case should_return_early {
    True -> ReturnedEarly
    False -> ReturnedAsPlanned
}
```

Using `bool.guard`, we can simplify this:

```gleam
let should_return_early = True

use <- bool.guard(should_return_early, ReturnedEarly)

ReturnedAsPlanned
```

The beauty of this is that it _feels_ like an early return. And sure, it only
works with bools but there are libraries that open that up too.

## The Ugliness of Case Statements

Aesthetically, gleam's weakness is the case statement.

Simply put, it's **ugly**.

Everyone seems to love gleam's pipes, `|>`, for composing function calls. To
cater for this, a lot of stdlib functions are pretty simple. In fact, a lot of
them simply wrap a case statement.

Case statements kinda suck. They don't feel very composable. Take an if
statement in gleam. Many FP langs have the `ifelse` sort of function. One can
write it in gleam like:

```gleam
fn ifelse(condition, true, false) {
    case condition {
        True -> true
        False -> false
    }
}

let num = -23
let positive_num = ifelse(num > 0, num, 0)
```

It's nice! But then again, all it does is replace a case statement. People (imo)
like the `ifelse`. I literally avoid checking boolean conditions because writing
an `ifelse` using a case sucks.

That's why I **abuse** use.

I love the [given library](https://github.com/inoas/gleam-given). But it is
hardly the _worst_ way to use `use`.

## Introduction to Abuse

First, let us get some of the simpler ones. A common useful one is `param` and
`defer`.

### param

`param` is used in a couple gleam libraries and its primary use case is for
making curried functions. Gleam doesn't natively auto-curry functions (you can a
smidge with function captures but it isn't the same). There used to be helpers
in the stdlib for making functions curried but they have long since been
removed. I think part of the issue was that you cannot make a simple function
that will curry any function. Instead, the stdlib had to make a curry function
for all possible function arities...which if we had some metaprogramming,
wouldn't have been a problem...

`param` looks like this:

```gleam
fn param(f: fn(a) -> b) -> fn(a) -> b {
  f
}
```

At first, `param` seems useless; it just returns the function it's given. But
when combined with `use`, it acts as a placeholder:

```gleam
let f: fn(List(a)) -> fn(fn(a) -> b) -> List(b) = {
    use list <- param
    use map <- param

    list.map(list, map)
}
```

`use list <- param` effectively says, 'I expect a parameter named list to be
provided here later.' It turns the `use` block into a function that is waiting
for that argument. Basically, it's a curried `list.map`.

I use this all the time. Mostly because gleam doesn't let you destructure tuples
and records in function parameters. For example:

```gleam
fn some_func(#(a, b): Tuple(some_type_a, some_type_b)) { todo }
```

This would not compile. Instead you would have to do this:

```gleam
fn some_func(tuple: Tuple(some_type_a, some_type_b)) {
    let #(a, b) = tuple
    todo
}
```

Which is kind of annoying. We can solve this with

```gleam
let some_func = {
    use #(a, b) <- tuple
    todo
}
```

### defer

The `defer` function is meant for side-effect code. Gleam doesn't have a good
way for typing side-effects (flix my beloved) but there are times where one
needs to do something at the end of function like clean up memory.

Take for example:

```gleam
let mem = allocate_some_memory_via_ffi(bytes: 32)

write_to_mem(mem:, value: "hello", offset: 0)

print_mem(mem:)

clean_up_mem(mem:)
```

Now, all of these are side-effect calls (and probably return `Nil`) but, as odin
and zig users learnt, deallocating where you allocate is generally a good idea.
We can do this by defining a `defer`:

```gleam
fn defer(deferred: fn() -> any, current: fn() -> result) -> result {
  let result = current()
  deferred()
  result
}
```

This function takes in the deferred code (i.e. `deferred`) which we call after
the current context, `current`. We can rewrite above as

```gleam
let mem = allocate_some_memory_via_ffi(bytes: 32)
use <- defer(fn() {clean_up_mem(mem:)})

write_to_mem(mem:, value: "hello", offset: 0)

print_mem(mem:)
```

The reason we wrap the deferred code in a function is to make it **lazy**.
Otherwise, we would call it immediately.

## Graduate Abuser

The `use` context _wraps_ context. Wouldn't be useful to _unwrap_?

Introducing, my friend, `apply_with`:

```gleam
fn apply_with(to a: fn(a) -> b, with b: fn() -> a) -> b {
  a(b())
}
```

This function runs in the inner context of `use` and then calls the `deferred`
function with result.

A bit of a word-salad but you can think of it as a clever trick that lets us
unwrap our expressions. For example:

```gleam
fn some_func(route) -> Response {
    let result: Result(View, PageError) = case route {
        "home" -> home(route)
        "projects" -> projects(route)
        _ -> Error(InvalidRoute)
    }

    use view: View <- result.map(result) // <- this is a type error!!!

    case view {
        Home -> "this my home page"
        _ -> "WIP"
    }
}
```

Obviously, this returns a `Result(String, PageError)`, not a `Response`.
Typically, we can quick-fix this by wrapping everything:

```gleam
fn some_func(route) -> Response {
    let body: Result(String, PageError) = {
        let result: Result(View, PageError) = case route {
            "home" -> home(route)
            "projects" -> projects(route)
            _ -> Error(InvalidRoute)
        }

        use view: View <- result.map(result)

        case view {
            Home -> "this my home page"
            _ -> "WIP"
        }
    }

    case body {
        Ok(body) -> Response(200, body)
        Error(page_error) -> Response(400, string.inspect(page_error))
    }
}
```

Which is...ew. Instead, let's use our `apply_with`:

```gleam
fn some_func(route) -> Response {
    use <- apply_with({
        use body: Result(String, PageError) <- param

        case body {
            Ok(body) -> Response(200, body)
            Error(page_error) -> Response(400, string.inspect(page_error))
        }
    })

    let result: Result(View, PageError) = case route {
        "home" -> home(route)
        "projects" -> projects(route)
        _ -> Error(InvalidRoute)
    }

    use view: View <- result.map(result)

    case view {
        Home -> "this my home page"
        _ -> "WIP"
    }
}
```

et voilà. You can we still get the niceness of using `use` with `result.map` but
instead of the function returning that, we unwrap it with our `apply_with`,
turning our `Result` into a `Response`. Admittedly, the order in which the code
runs is. The `use` syntax tends to wraps types as we go _down_ the function but
our `apply_with` does the opposite; it unwraps as we go _up_. You can
mix-and-match a whole lot of these to cause maximal psychic damage to gleam
noobies.

## Professional Abuser

Recently, I have been building an UI library for a gleam mobile app. The problem
I had is that the only UI API I know of for gleam is
[lustre](https://github.com/lustre-labs/lustre)'s

```gleam
fn view(model) {
  let count = int.to_string(model)

  div([], [
    button([on_click(Incr)], [text(" + ")]),
    p([], [text(count)]),
    button([on_click(Decr)], [text(" - ")])
  ])
}
```

(from their readme)

And, to be honest, I hate it. It's how people recommend doing webdev in Ocaml as
well and I hated it then too. In fact, I setup `dune` to compile
[reasonml](https://reasonml.github.io/) with a JSX preprocessor just to have
something more HTML-like.

Unfortunately, gleam doesn't have anything preprocessor like (i.e.
metaprogramming). Some libraries get around this by shipping an application that
will do preprocessing for you, like turning SQL files into gleam types with
decoders.

Instead, I decided to go full on the other way. No need to conform to one-to-one
mapping with the web. The lustre example above shows how to do 2 buttons with
text in the middle. My UI library is like so:

```gleam
fn view(model) {
    let count = int.to_string(model)

    use append, done <- list

    use <- append({
        use <- has(button(Incr))
        text(" + ")
    })

    use <- append(text(count))

    use <- append({
        use <- has(button(Decr))
        text(" - ")
    })

    done
}
```

The reason I like this is because it isn't as nested but I still keep the listed
things on the same context level. The `done` value stops the user from mixing UI
element specific functions, like the `has` function, in the top level scope.

Another thing is attributes. In lustre, it is a list that the user defines, the
first parameter passed to all the elements in the example. My UI library is done
like so:

```gleam
use <- has(button(ChangeView(Tower)))
use <- font_size(20)
use <- color(Background, color.green)
text("Tower")
```

This makes a button with the text "Tower", have a font_size of 20 and a
background color of green. You can even apply it for things like orientation and
gap:

```gleam
use <- gap(10.0)
use <- orient(Horizontal)
use <- pad(All, 10.0)
use <- color(Background, color.blue)

use append, done <- list

use <- append({
use <- has(button(ChangeView(init())))
    use <- font_size(20)
    use <- color(Background, color.red)
    text("Rest Area")
})

use <- append({
    use <- has(button(ChangeView(Tower)))
    use <- font_size(20)
    use <- color(Background, color.green)
    text("Tower")
})

done
```

The `list` function I introduce looks like

```gleam
pub opaque type Done(msg) {
  Done(List(Element(msg)))
}

pub fn list(
  f: fn(fn(Element(msg), fn() -> Done(msg)) -> Done(msg), Done(msg)) ->
    Done(msg),
) -> element.Element(msg) {
  let Done(elements) = {
    use first, done <- f(_, Done([]))
    let Done(rest) = done()
    Done([first, ..rest])
  }
  group(elements)
}

pub fn group(elements children: List(Element(msg))) {
  element.Element(
    tag: "group",
    attributes: attribute.new(),
    on: option.None,
    children: element.Many(children),
  )
}
```

What really makes it complex is returning another callback function...in the
callback function.

Lets break it down real quick. The `list` function takes a single callback, `f`.
This callback has two parameters: `append` and `done`. `append` is a function
that takes in the element we want to append and a callback.

The callback expects the user to return a `Done` type. The user can do this in
two ways. One, they use `append` again, as its return type is also `Done`. This
lets us nest our `append` calls. Or two, we simply return the `done` we were
provided by the `list` function, which is also the only way to end the `list`
`use` context.

## A Forewarning

The main problem with `use` is that it confuses readers. Of course, with time,
it becomes easier to parse and eventually you start making incredibly cursed
functions to fulfill your perverse desires. Even my own code scares my friends
away:

```gleam
let current_ids =
    option.map(object, {
        use object: render.Object <- util.param
        let id = object.id

        use _ <- function.tap(set.insert(current_ids, id))

        use <- bool.guard(cache.has_checksum(object:), Nil)

        use <- util.defer(fn() { cache.set(object:) })

        use <- bool.guard(cache.has(id:), Nil)

        render.create(object:)
    })
    |> option.unwrap(current_ids)
```

...yeah. This has a lot of mutable state. Perhaps one day I will figure out a
nice way to type guard it, like regions in other immutable langs.

I wish more languages had `use`. Although, I imagine that if
[Louis](https://lpil.uk/) ever saw my code, he would nuke this feature in a
heartbeat...
