GAS: Game Animation System
==========================
The game animation system (GAS) animates objects for interactive media of all sorts including video games and other motion graphics.

The author's intent is to translate GAS into multiple languages as a learning exercise.  Right now, only GoLang is provided.

GAS herein is not production ready. Nevertheless, it is a place to start and extend. Custom animations systems can often do more with less, so you might want to experiment.

![GAS Stillshot](https://raw.githubusercontent.com/jkassis/gas/main/screens/frogger.intro.png)

Documentation
-------------
No formal documentation other than comments, source code, and the example Frogger Game Welcome Screen.

The API is mostly self-explanatory, though inspection of the example Frogger anim code will yield the most rapid understanding.


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
The API is deliberately terse and intended to produce smooth reading animation sequence code. Normally, I prefer fully spelled out variables that emphasize data hierarchies and fn names in object-verb order. eg. orderItemPut(item: Item). Think "dot-notation" without the dots.

I feel this strategy helps...
  * control entropy since I only need to "think" in one order (dot-order).
  * improve maintainability... because remembering the abbreviation over time becomes difficult.

I like to avoid use of the shift key, so snake_case and PascalCase are right out for me. Kebab-case isn't bad, but I prefer camelCase to keep things tight.

That said, I'm flexible to whatever the team chooses and prefer opinionated languages like GoLang that strictly enforce style.


Known Issues
-----------------------

GoLang
* Datatypes  
  excessive casting and use of int64 creating a large memory footprint for display objects.
* Error Handling  
  The SDL layer surfaces many errors which GAS mostly discards. This is one of the reasons gas is not production ready.
* Public API  
  Little thought into what fields should allow public access
* Recycling  
  This systems should recycle dobs to avoid garbage collection. Currently no pooling of textures, dobs, etc.
* z-layering  
  Controlled by OrderedMaps, which I have not benchmarked, tested thoroughly.


## License
No public license to use GAS, the game animation system. All rights reserved ©2023 Jeremy Kassis.

Private licenses available on request.



