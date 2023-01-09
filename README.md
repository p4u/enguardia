# enguardia
Quick&Dirty codi amb Go que fa un scraping dels audios del programa "En Guàrdia" de Catalunya Radio (CCMA), els descarrega en local i serveix una web neta amb tot el contingut. Inclou un cercador per poder filtrar per títol o descripció.

Pots llançar el programari amb docker compose:
```
docker-compose build
docker-compose up -d
```

Els capitols es descarregaran al directori "data". Un cop finalitzat, la web estarà disponible a http://localhost:8080
Cada cop que es llanci de nou el procès, es descarregaran els capitols nous que encara no han sigut descarregats.

Es pot veure un exemple de la web a http://enguardia.dabax.net (no asseguro la seva disponibilitat, podria estar caiguda).
