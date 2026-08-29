# macOS Installation

Required to have [Homebrew](https://brew.sh) installed.

The macOS branch out of goro branch, which contains many configurations required for goro to work with rathena.

Check the other configuration files for more details.

Clone and checkout to macOS+Goro branch:
```sh
git clone https://github.com/jacksonbenete/rathena.git
cd rathena
git checkout macos
git pull --ff-only
```

Install dependencies and compile:
```sh
brew install mariadb
brew install pcre
./configure --with-mysql=/opt/homebrew/opt/mariadb/bin/mysql_config --with-pcre=/opt/homebrew/opt/pcre
make clean
make server
```

Setup and start the database for the first time (update -v parameters with your chosen path):
```sh
container system start
container volume create rathenadb
container run --name=rathenadb -e MARIADB_ROOT_PASSWORD=root \
  -p 3306:3306 \
  -v "~/code/rathena/sql-files:/docker-entrypoint-initdb.d:ro" \
  -v rathenadb:/var/lib/mysql \
  mariadb
```

Next time you just need to restart the container and your database data is persisted as long as you keep the volume defined previously:
```sh
container start rathenadb
```

In `rathena/conf/grf-files.txt` should be pointing to `~/` dir in rathena macos branch, make sure the OldRO data is on `~/OldRO` or `/Users/youruser/OldRO`, or provide the correct path.

Make sure the `<connection>` node in `OldRO/data/clientinfo.xml` points to `127.0.0.1` if you're running locally.
If there's more than two connection nodes, remove one and let only the localhost or 127.0.0.1.

Start the three rAthena servers from separate terminals:
```sh
./login-server
./char-server
./map-server
```

Compile and run:
```sh
clear && CGO_ENABLED=0 go run -tags nofakecgo . --data-dir ~/OldRO
```

It might be necessary to change the initial map and coordinates on the `char` database in the case your first login seems broken in a deep black screen.

To make a test account a GM:
```sql
UPDATE login SET group_id = 99 WHERE userid = 'youruser';
```

It might be the case that eventually you must try to recompile mapcache in rAthena.
For reference:
```sh
make tools
./mapcache \
  -grf conf/grf-files.txt \
  -list map_index.txt \
  -cache db/pre-re/map_cache.dat \
  -rebuild
```
