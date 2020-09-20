How pairings work in Excalibur
==============================

I've tried to follow [FIDE's basic Swiss pairings rules for chess](https://handbook.fide.com/chapter/C0401) in most things. I've tried to follow the Netrunner tournament rules wherever they don't conflict with the FIDE rules.

The pairings engine essentially looks at every possible pairing and picks the best one. If there are multiple equally good pairings, it picks one at random, though it's possible that not all best pairings are equally likely to be chosen (randomness is hard).

In practice, it discards some pairings before they're fully worked out if it can already tell they won't be the best.

(Thanks to SpaceHonk for his posts [in this old reddit thread on how NRTM does this](https://www.reddit.com/r/Netrunner/comments/2tni66/nrtm_swiss_pairing_algorithm/), which set me on the right track.)

Comparing pairings
------------------

When comparing two pairings to decide which one is better, the engine looks at the following things:

* If one has fewer rematches, the one with fewer rematches is preferred. Since this is the highest priority rule, unless you have as many rounds as players, rematches should always be prevented.

  * In the weird and unlikely case where one pairing involves some players facing each other for a _third_ time, any pairing that involves fewer people facing for a third time is preferred, even if the alternative has more people facing the same opponent a second time.
  
    This continues for higher (and even more unlikely) levels of rematches. Preventing fourth matches between the same players is more important than preventing third matches between the same players, preventing fifth matches is more important than preventing fourth matches, etc.
    
  * Repeat byes for the same player are also prevented here.
    
* If they're equally good as far as rematches go, then if they both have byes, the one where the player with the bye has the lower score is preferred.

* If they're equally good for byes too, then for each pairing, it looks at the difference between how many times each player would have played Corp and how many times they would have played Runner after that pairing. The pairing for which this side difference is three or more for fewer players is preferred.

  * This is still a high priority criteria, and it's expected that zero players will have a side difference of three or more at any point in most tournaments.
  
  * Preventing side differences of four is more important than preventing side differences of three, preventing differences of five is more important than preventing differences of four, etc.

* If they're still equally good, then for each pairing, it looks at how many times in a row each player would have played the same side after that pairing. The pairing for which this streak length is three or more for fewer players is preferred.

  * This is still a high priority criteria, and it's expected that zero players will have any streaks of three or more in most tournaments.

  * Preventing streaks of four is more important than preventing streaks of three, preventing streaks of five is more important than preventing streaks of four, etc.
  
* Now it looks at score groups. It looks at how many matches in each pairing are between players in different score groups, and how far apart those score groups are.

  * A score group is all the players with the same number of prestige points.
  
  * Preventing matches that cross more score groups is more important than preventing matches that cross fewer score groups, or that are between players in adjacent score groups.
  
* If the pairings are equally good for score groups too, it tries to pick the one where fewer players have a side difference of two. Side differences of more than two were avoided above.

* If the pairings are still equally good, it tries to pick the one where the most players are not playing the same side as the previous round.

* If there's still no difference, then the pairings are equally good in every way as far as the pairings engine is conceerned. It picks whichever one it found first.

Notes
-----

The FIDE rules permit side differences and streaks of three in the last round only. Excalibur doesn't currently have special handling for the last round, but it might be a good idea to add at some point, in order to prioritize matching by score group in the final round.

I also haven't implemented this part of the FIDE rules:
> A player who has already received a pairing-allocated bye, or has already scored a (forfeit) win due to an opponent not appearing in time, shall not receive the pairing-allocated bye.
