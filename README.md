Excalibur is a program for running Netrunner tournaments with single-game rounds. It is not very polished, but it should be adequate for running small test tournaments to experiment with the single-game-round tournament format.

You can read a rambling explanation of how the pairings rules work in [pairings.md](pairings.md).

Instructions
============

1. On the command line, run Excalibur with the name of a save file, like so:

        excalibur test_tournament

    If the specified file exists, the newest save in it is loaded; if it doesn't exist, it is created. If it exists but is not an Excalibur save file, probably bad things happen. I haven't tried it.

2. Go to http://localhost:8080/ in your browser.
