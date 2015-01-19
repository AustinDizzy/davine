#![Davine](https://raw.githubusercontent.com/AustinDizzy/davine-share/master/assets/Davine-text-render.png "Davine - Plant your vines, we'll watch them grow.") 
[![Build Status](https://travis-ci.org/AustinDizzy/davine.svg)](https://travis-ci.org/AustinDizzy/davine)


[Davine](//www.davine.co) is an open data, open source, and ad-free social analytics service built for, and in no way affiliated with, [Vine](https://vine.co).


Davine tracks the following from each user it crawls or is manually introduced to:
  - Total Loop Count
  - Total Followers/Following
  - Original Post and Revine counts
  - Previous usernames, descriptions, and profile locations
  - Verified on, private on, and explicit on date(s)
  - [and more](app/types.go)

### Built With
  - [Go](http://golang.org) - a highly efficient and scalable language
  - [Google App Engine](https://cloud.google.com/appengine) - A perfect match for Go; a highly scalable and available PaaS
  - [Google Cloud Datastore](https://cloud.google.com/datastore/) - A fast and highly replicatable database solution by Google
  - [Highcharts](http://highcharts.com) - An easy to use and highly customizable charting solution.
  - [Bootstrap](http://getbootstrap.com) - An expansive boilerplate, enabling fast mockups and templates
  - [hoisie/Mustache](https://github.com/hoisie/mustache) - A Go port of the [mustache](http://mustache.github.io) templating spec
  - [btmills/GeoPattern](https://github.com/btmills/geopattern) - A JS port of [jasonlong/geo_pattern](https://github.com/jasonlong/geo_pattern), an entropic SVG pattern generator.
  - [Vine](https://vine.co) - Kinda sorta. A social network where users can share six seconds of looping video.

### Cron Schedule(s)

Every 3 hours, Davine automatically stores any user records that are on the first three pages of Vine's "[Popular Now](https://vine.co/popular-now)" page. As the Davine project progresses in development and popularity, we will begin to crawl other Vine pages such as Vine channels, playlists, and Editor's Picks.

Every 24 hours (may change as development and time goes on), Davine pulls all user information for users it has discovered -- whether crawled from the popular page(s) or introduced manually by another user -- and appends their statistics in a user record stored and highly replicated on Google Datastore. Public datasets will soon be available for download and browsing free of charge.

Davine will begin crawling Vine on December 31st, 2014.

License
----

Davine is licensed under the terms of the MIT License. Full text of the MIT License can be found in the included [LICENSE](LICENSE) file.
