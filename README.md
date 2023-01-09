# enguardia
Programa que fa un scraping dels audios del programa "En Guàrdia" de Catalunya Radio (CCMA) i serveix una web neta amb tot el contingut. Inclou un cercador per poder filtrar per títol o descripció.

Per iniciar el programa amb docker compose:
```
docker-compose build
docker-compose up -d
```

Els capitols es descarregaran al directori "data". Un cop finalitzat, la web estarà disponible a http://localhost:8080
