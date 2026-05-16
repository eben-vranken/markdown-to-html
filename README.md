# 0.2 — Markdown to HTML Converter

A from-scratch Markdown parser written in Go. It reads a `.md` file from `in/`, converts it to HTML, and writes the result to `out/`. No regex, no parsing library — just `strings`, `bufio`, and a handful of state flags.

This is project **0.2** of my Go learning track — the second step on a self-paced parallel to my Rust backend roadmap.

## How it works

The parser has two distinct layers, because Markdown itself has two distinct layers:

- **Block elements** are determined by what a whole line *is*: a heading, a paragraph, a horizontal rule, a list item, a blockquote, a code fence. These are detected by prefix checks (`#`, `>`, `-`, `1.`) or by exact match (`---`, ` ``` `).
- **Inline elements** live *inside* a line: `**bold**`, `*italic*`, `` `code` ``, `[link](url)`, `![alt](url)`. These don't care about line structure — they're string replacements within content that block parsing has already extracted.

`processLine` handles blocks; `parseInline` handles everything within.

### Block parsing is a state machine

A paragraph isn't a single line — it's a *group* of lines bounded by blank lines or other block elements. Same for lists and code fences. So `processLine` carries four booleans across iterations:

```go
inParagraph    *bool
inUnorderedList *bool
inOrderedList  *bool
inCodeBlock    *bool
```

The pointers matter: each call to `processLine` needs to *mutate* these flags so the next call sees the updated state. Without pointers, `processLine` would only modify its local copy and the caller's variables would never change.

The flags only flip at transitions. A plain text line entering a paragraph emits `<p>` and flips `inParagraph` to true; subsequent plain text lines just emit content. A blank line or heading flips it back to false and emits `</p>`. The same shape applies to lists.

### Lists need a recursive trick

When a non-list line appears right after a list (no blank line between them), the list needs to close *and* the current line needs to be processed normally. A simple `else if *inUnorderedList { ... }` branch emits `</ul>` but then has nowhere to handle the closing line. The fix is to call `processLine` recursively after closing the list:

```go
} else if *inUnorderedList {
    *inUnorderedList = false
    fmt.Fprintln(outputFile, "</ul>")
    processLine(line, inParagraph, inUnorderedList, inOrderedList, inCodeBlock, outputFile)
}
```

This isn't infinite recursion because `*inUnorderedList` is already false by the time the recursive call happens — the same branch can't re-fire.

### Code blocks bypass everything else

The first thing `processLine` does is check `*inCodeBlock && !startsWithFence(line)`. If so, the line is written verbatim and the function returns. No heading checks, no paragraph wrapping, no inline parsing on the contents. Only the ` ``` ` fence itself toggles the flag and emits the wrapping `<pre><code>` / `</code></pre>` tags.

### Inline parsing is ordered most-specific-first

`parseInline` runs a series of `strings.Replace` loops, replacing pairs of delimiters alternately (first match → opening tag, next match → closing tag). The catch is that several inline patterns share characters: `***bold italic***` contains `**bold**` which contains `*italic*`. If you replace `*` first, you obliterate the longer patterns before they can match. So the loops run from most-specific to least:

1. `` ` `` → `<code>`
2. `***` → `<strong><em>`
3. `**` → `<strong>`
4. `*` → `<em>`
5. `![alt](url)` → `<img>`
6. `[text](url)` → `<a>`

Images come before links for the same reason: `[text](url)` is a substring of `![text](url)`.

## How to run it

Put one or more `.md` files into `in/`. Then:

```sh
go run .
```

The program lists the files it found and prompts for the index of the one to convert. The result is written to `out/<name>.html`.

```
Files found:
0: index.md
File to generate: 0
Generating file: out/index.html
```

## What I learned

- **State that outlives one iteration must outlive the loop.** Declaring `inParagraph` *inside* the `for` body resets it every iteration, defeating the point. The fix — moving it above `for {` — is a tiny change but the conceptual lesson (where you declare a variable defines what remembers what) shows up everywhere.
- **Strings are immutable in Go.** `strings.Replace(s, ...)` returns a new string; it doesn't mutate `s`. Forgetting to assign the result back is an instant infinite loop, because `strings.Contains` keeps returning true on the unchanged value.
- **Pointers cross function boundaries for mutation.** Go passes arguments by value, so a `bool` parameter is a *copy*. To let a helper change the caller's flag, the parameter has to be `*bool` and dereferenced inside. Same idea as a C pointer, much friendlier syntax.
- **Prefix-matching order matters.** Checking `strings.HasPrefix(line, "#")` before `"######"` would route every h6 line to the h1 branch. The general rule: when patterns nest, check the longest first. This applies to headings *and* to inline `***`/`**`/`*` — same principle, two contexts.
- **Bounds checks before indexing.** `line[0]` panics if the line is empty. The ordered-list detection needs `len(trimmed) >= 2 && unicode.IsDigit(rune(trimmed[0])) && trimmed[1] == '.'`, in that order, so Go's short-circuit evaluation keeps the index access from ever running on a too-short line.
- **`byte` vs `rune`.** Indexing into a string with `[i]` gives you a `byte` (uint8). `unicode.IsDigit` expects a `rune` (int32). Go doesn't auto-convert between numeric types, so you have to write `rune(s[0])` explicitly. The byte/rune distinction is the same one that bit me on character counting in 0.1, just in a new outfit.
- **Recursion is fine when the state guarantees termination.** Calling `processLine` from inside itself sounds dangerous, but if you flip the only flag that could send you back into that branch *before* the recursive call, the recursion is bounded — at most one level deep per line.
- **Format functions don't all behave the same.** `fmt.Fprintln(w, "<li>", content, "</li>")` inserts spaces between the arguments (`<li> content </li>`). For tight HTML you want `fmt.Fprintln(w, "<li>"+content+"</li>")` — single string, no surprise whitespace.
- **`fmt.Fprint` takes an `io.Writer`, which means anything writable.** Threading `outputFile *os.File` through `processLine` worked, but the more idiomatic type would be `io.Writer` — the function only cares that the destination implements `Write`, not that it's specifically a file. Same lesson as 0.1's "accept interfaces."
- **The two-pass mental model generalizes.** Block-then-inline is the same shape every Markdown parser uses, including CommonMark. Building it badly from scratch made the well-known good versions easier to read afterward.
