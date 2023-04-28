GAS: Game Animation System
==========================
The game animation system (GAS) animates objects for interactive media of all sorts including video games and other motion graphics.

I intend to translate GAS into multiple languages as a learning exercise.  Right now, I only provide GoLang.

I would not recommend GAS herein for production. Nevertheless, this repo can serve as a good place to start and extend. Custom animation systems can often do more with less, so you might want to experiment.

![GAS Stillshot](https://raw.githubusercontent.com/jkassis/gas/main/screens/frogger.intro.2.png)

Documentation
-------------
No formal documentation other than comments, source code, and the example Frogger Welcome Screen.

The API mostly explains itself, though inspection of the example code will yield the most rapid understanding.


Installation
------------
Pre-Requisites:
See https://github.com/veandco/go-sdl2

Code:
```
git clone https://github.com/jkassis/gas
cd gas/go
go run main.go
```

Style
-------------
Deliberately, the API may seem terse, but it should afford smooth reading animation sequences. Normally, I prefer fully spelled out variables that emphasize data hierarchies and fn names in object-verb order. eg. orderItemPut(item: Item). Think "dot-notation" without the dots.

I feel this strategy helps...
  * control entropy since I only need to "think" in one order (dot-order).
  * improve maintainability... because remembering the abbreviation over time becomes difficult.

I like to avoid use of the shift key, so snake_case and PascalCase are right out for me. Kebab-case... not bad, but I prefer camelCase to keep things tight.

That said, I'm flexible to whatever the team chooses and prefer opinionated languages like GoLang that strictly enforce style.

![GAS Stillshot](https://raw.githubusercontent.com/jkassis/gas/main/screens/frogger.intro.1.png)

Ports
-----------------------
### GoLang
#### Re: Building for Web and WASM
Go compiles to multiple platforms and architectures and allows use of CLibs using CGO bindings... which makes this project possible. GAS for GoLang renders through CGO bindings to platform-native builds of the SDL libraries. But getting GAS in GoLang to run on the web gets complicated...

*Native Toolchain*
The native GoLang toolchain can build a WASM target using something like `GOOS=js GOARCH=wasm go build -o main.wasm`. But it does not support CGO bindings to platform libraries for the wasm target since a WASM app runs in a sandbox that does not provide access to those libs. The native toolchain could at least support builds with staticly included CLibs, but this only makes sense if the CLibs are *portable*. The Go Maintainers have not promised this support, so port your portable code to GoLang first if you want to build it into a WASM target.

*Emscripten*
Googling for "GoLang SDL WASM" surfaces [EMScripten](https://emscripten.org/docs/introducing_emscripten/about_emscripten.html), a toolchain for compiling various languages to WASM. It has similar limitations... it can only build *portable* code and libraries to WASM targets as described in [Emscripten Porting](https://emscripten.org/docs/porting/guidelines/index.html).

And yet... Emscripten *does* provide support for building applications that use the SDL2 CLibs.

How? The Emscripten toolchain replaces CLang calls to the CLang SDL libs with WASM calls to a Javascript port of SDL that renders with WebGL.

See [*Javascript* port of SDL1.2](https://github.com/emscripten-core/emscripten/blob/main/src/library_sdl.js). That’s from-scratch code and some things aren’t doable in a web browser. It’s a "Good Enough" implementation of the API, and older games might just work.

For SDL2, the actual SDL2 codebase is ported to Emscripten, so you can literally just compile it and link it to your project. See [SDL and WebAssembly](https://discourse.libsdl.org/t/sdl-and-webassembly/24611/5). Ultimately, the OpenGL calls get converted to WebGL calls.

Why? SDL essentially defines abstrations for rendering, sounds, windows, etc. The Web API provides this too. So the Web API and the SDL API occupy roughly the same tier in the front-end stack. Why not use the Web API? Your stack probably should. But your code needs an abstraction layer to keep your game code clean. That's GAS.

*Emscripten and GoLang*
Emscripten only compiles LLVM-based languages. GoLang has a custom compiler toolchain. And [GoLLVM](https://go.googlesource.com/gollvm/) is not production-ready.

*See also...*
* [WASM Spec Summary](https://webassembly.github.io/spec/)
* [Mozilla Intro to WASM Concepts](https://developer.mozilla.org/en-US/docs/WebAssembly/Concepts)
* [The WASM spec](https://github.com/WebAssembly/spec/)
* [Using WebAssembly to call Web API methods](https://stackoverflow.com/questions/40904053/using-webassembly-to-call-web-api-methods).

*GopherJS*
GopherJS compiles GoLang code to javascript. It does not use the GoLang compiler / toolchain. It provides a separate compiler that generates Javascript, not a binary executable. See [Creating WebGL apps with Go](https://blog.gopheracademy.com/advent-2018/go-webgl/) for a story about using GopherJS with three.js to do 3D rendering.

While people rave about GopherJS, your Go code will never look the same. You will be writing Javascript in Go. Maybe you should just write Typescript? And WASM intends to improve the performance .js, so why target .js?

By itself, WebAssembly cannot currently directly access the DOM; it can only call JavaScript, passing in integer and floating point primitive data types. Thus, to access any Web API, WebAssembly needs to call out to JavaScript, which then makes the Web API call. So, if WASM app calls out to the Web APIs frequently, it goes through .js, which can impact performance.

Putting a finer point on it... GopherJS provides a binding interface and [bindings to popular .js packages](https://github.com/gopherjs/gopherjs/wiki/bindings). When GopherJS compiles the GoLang code it produces .js that calls .js through the generic binding interface.

If you build to WASM with the native GoLang toolchain, your GoLang to .js bindings flow through the generic GoLang to .js binding interface... [syscall/js](https://cs.opensource.google/go/go/+/master:src/syscall/fs_js.go).

The picture shaping up here... WASM can optimize compute intensive operations, but might not provide performance advantage when used to drive the display layer, which requires high frequency WebGL calls that must flow through / originate from .js.

But if/when WASM apps get direct access to WebAPIs that ultimately bypass .js... the table flips. [WASI](https://github.com/bytecodealliance/wasmtime/blob/main/docs/WASI-intro.md), the Web Assembly System Interface, represents some progress in this direction. TinyGo leads the way in WASI support for GoLang. See [WASI Hellow World](https://wasmbyexample.dev/examples/wasi-hello-world/wasi-hello-world.go.en-us.html).

Whether you compile your GoLang code to .js or WASM, the GoLang code will call into GoLang bindings. And someone needs to write these bindings... to Web APIs and to native Javascript libraries. Perhaps the GoLang to Web crowd has grown overly enthusiastic with the success of Gopher.js? But perhaps the ecosystem doesn't need more?


#### Known Issues
* Datatypes
  excessive casting and use of int64 creating a large memory footprint for display objects.
* Error Handling
  The SDL layer surfaces many errors which GAS mostly discards. Once again... not production ready.
* Public API
  Little thought into what fields should allow public access
* Recycling
  This systems should recycle dobs to avoid garbage collection. Currently no pooling of textures, dobs, etc.
* z-layering
  Controlled by OrderedMaps, which I have not benchmarked, tested thoroughly.
* fn calls
  Ans implement an abstract interface so that Dobs can run generic Ans. This means we have to access properties through accessors. Should be possible to fix this with generics and inverting the embed. Instead of embedding BaseAnim in each An, we would embed the specific An inside base.
  It should also be possible with this fix to init startTick for each An before the first invocation of the start tick to remove some of the boilerplate code from the An.Tick fns.
* memory leak
  looks like there is a huge memory leak that i need to shake down.


License
-------
No public license to use GAS, the game animation system. All rights reserved ©2023 Jeremy Kassis.

Private licenses available on request.




See Also
--------

### GoLang for Cross-Platform Game Engines
https://blog.gopheracademy.com/advent-2018/go-webgl/
https://github.com/soypat/threejs-golang-example
