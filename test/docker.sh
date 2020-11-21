#!/bin/sh

base="https://pass.bolt"

root_password="asdf1234"
mariadb_database="thedb"
mariadb_user="theuser"
mariadb_database="thepassword"

docker run --rm --name mariadb -d \
             -e MYSQL_ROOT_PASSWORD=${root_password} \
             -e MYSQL_DATABASE=${mariadb_database} \
             -e MYSQL_USER=${mariadb_user} \
             -e MYSQL_PASSWORD=${mariadb_password} \
             mariadb

sleep 2
mariadb_container_host="$(docker inspect --format '{{ .NetworkSettings.IPAddress }}' mariadb)"

docker run --rm --name passbolt \
             -p 80:80 \
             -p 443:443 \
             -e DATASOURCES_DEFAULT_HOST=${mariadb_container_host} \
             -e DATASOURCES_DEFAULT_PASSWORD=${root_password} \
             -e DATASOURCES_DEFAULT_USERNAME=root \
             -e DATASOURCES_DEFAULT_DATABASE=${mariadb_database} \
             -e APP_FULL_BASE_URL=${base} \
             passbolt/passbolt

echo Running DB at $mariadb_container_host and PB at $base

#docker exec passbolt su -m -c "/var/www/passbolt/bin/cake passbolt register_user -u your@email.com -f yourname -l surname -r admin" -s /bin/sh www-data
