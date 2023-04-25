# GAS: Game Animation System
# Fueling Frogger[![License: CC0-1.0](https://img.shields.io/badge/License-CC0_1.0-lightgrey.svg)](https://spdx.org/licenses/CC-PDDC.html)

Multiple implementations of a simple Frogger game in many languages.

![Frogger gameplay](https://raw.githubusercontent.com/jkassis/frogger/main/screenshots/play.png)

Frogger is simple enough to write in xxx lines, yet requires fluency with many different language and app fundamentals...

* Workflow
  * Installation and setup
  * Inner Loop: Run, Debug, Change
* Design
  * Object-Orientated Design
  * Finite State-Machines
  * Determinism & Non-Determinism
  * Multi-Dimensional Data Structures
  * Physical Simulation
    * physical space
    * clock tick vs frame rate
    * velocity & collision
* IO
  * Video
    * Text-Mode
    * Graphics
      * Bitmaps and Textures
      * Double-Buffered Animation
      * Font rendering
  * Audio
    * Sound files and playback
  * Data serialization with JSON
  * Block Storage
    * R for Configuration
    * RW for Leaderboards
* Packaging & Distribution
  * Command Line Argument Processing
  * Cross-platform and architecture distribution

The following languages are implemented:

* Go (SDL)

## Terminology
As an artifact of interactive media, we will use theatrical terms to describe the Game.

As a simulation of a physical universe, we will use scientific and engineering terms of physics to describe the Game.

Finally, as a simulation of context in which intelligent agents make informed, critical decisions, we will use the terms of AI to describe the Game.


## The Stage (aka Space)
A 3D grid of element-id lists. In this case, a basic 16x16x16 grid.

While frogger seems at first like a 2D game, it is one of the first 3D games in which game pieces (frogs and logs) can occupy the same 2D grid space when viewed orthogonally from above.

While the rules of frogger do not allow game pieces to occupy the same grid cell in xyz space, we allow the grid to track superposition of game pieces. Thus we use a 3 coordinate system mapped to a list.


## The Set (aka Terrain)
This represents the lowest, immutable portion of the grid... we can call it the ground. Terrain is encoded as 2D ASCii art making it easy to visualize with the following pixels...

- O: Open
- x: Obstructed
- w: SloW
- !: Dangerous

Terrain in frogger is non-deformable, but easily could be.

## The Actors (aka Agents)
Game pieces (we will call them "Actors") occupy space on the map at arbitrary positions that can mutate over time. They can occupy arbitrary, non-contiguous grid cells (ie they can have non-contiguous shapes).

They are encoded as an array of json objects. The JSON object includes a type field to identify the type of agent and additional information unique to the type.

In this version of Frogger, we have the following Actors:

- Frog
- Log
- Car
- Truck
- Bus
- Shark

## The Context
Actors have context. Context is Game State that Actors can use to compute and mutate their own state.

In Frogger, frogs know they are on logs, for example. Frogs also know if their game piece overlaps another game piece (collision).

## The Players and Controls
Players control various Actors by sending them commands. In the general case, Players can control more than one actor by first selecting the actor and then sending it a command. In this version of Frogger, players command only one frog.

They can tell the Frog to move forward, backward, left, or right using the asdw keys.

## The Simulation
With each tick of the simulation, Actors can change position or shape.

Each Actor has a "tick" function that allows it to inspect it's internal and contextual state and mutate for the next frame.

The "tick" function includes a duration component to enable mutations that rely on the passing of game time, like velociy or momentum.

The tick function of each Actor gets called once per frame in non-deterministic order, which partially motivates the next design pattern.

## Messaging and Collisions
Agents can message each other during the tick. Messages are handled synchronously and because the simulation calls the tick function of Actors in non-deterministic order, the design must support bi-directional messaging.

For example, in this Frogger, we use messaging to signal collision. A log can collide with a Frog. A bus can collide with a Frog.  In the Frog's tick, it might move into an occupied space. It might also move into an unoccupied space that a bus or a log occupies later in the frame.

## Non-Determinism
Traffic is generated randomly with random velocities using a curved distribution model.

## Game Over Conditions
The game-over condition is reached when the Frog reaches the Goal Zone or collides with a dangerous object.

Once the game-over condition is reached, the simulation ends and the game loop continues.

The game-over melody and message are presented and the screen is updated one last time.

## Scoring
The player is scored for the number of boards cleared.

## Game Loop
The game loop...
- render the leaderboard, allow player to start game
- inits a fresh board
- simulates until game over
- detects win/loss
- on win... adds to the score and auto starts the next game
- on lose...
  - if the player made it to the leaderboard
    - play the success sequence
    - prompt for initials
    - re-render and wait for dismissal
  - if the player did not make it to the leaderboard
    - play the fail sequence
    - wait for dismissal



## Leaderboard
  - loads from disk and persists to disk
  - gets the filename from CLI argument


## Technical Details
### Screen Updates
Updates to the screen use double buffering: Updates are written to the frame-buffer and then flipped with the current buffer to update the screen in one go per frame.

The placement on the screen is done relative to the block size. The PNG images of the 7 colored blocks are 32x32 pixels. For implementations using bitmaps, this is a fixed screen size and placement. For implementations using GPU textures, the block_size constant can be changed to change the size and scale of the window.

Drawing the screen each frame performs the following steps:

* Clear
  Fill with a black rectangle (if needed per implementation)
* Draw each layer
  Draw Terrain, then actors from the bottom up.
* Draw the HUD (Head's Up Display)
  - Draw the wall separator
  - Print the logo
  - Write the lines and level
  - Show the next piece

### Keyboard Events
Where possible, the simulation loop polls for async io keyboard events.

Where not possible, the simulation loop runs at 1khz to gather input and ticks at the simulation frequency.

# See Also
- https://www.cs.swarthmore.edu/~meeden/cs81/s14/papers/DavisJake.pdf
- https://ceur-ws.org/Vol-2215/paper_3.pdf