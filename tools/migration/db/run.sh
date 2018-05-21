#!/bin/bash
# Copyright 2017 VMware, Inc. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

set -e

ISMYSQL=false
ISPGSQL=false
NOTARYDB=/notary-db
UPNOTARY=false

if [ "$(ls -A /var/lib/mysql)" ]; then
    # use the logic to handle run upgrade from less than 1.5.0 twice.
    if [ -e '/var/lib/mysql/PG_VERSION' ]; then
        ISPGSQL=true
    elif [ -e '/var/lib/mysql/created_in_mariadb.flag' ]; then
        ISMYSQL=true
    fi
fi

if [ "$(ls -A /notary-db)" ]; then
    UPNOTARY=true
fi

if [ "$(ls -A /var/lib/postgresql/data)" ]; then
    ISPGSQL=true
fi
if [ $ISMYSQL == false ] && [ $ISPGSQL == false ]; then
    echo "Please make sure to mount the correct the data volumn."
    exit 1
fi
if [ -z "$DB_USR" -o -z "$DB_PWD" ]; then
    echo "DB_USR or DB_PWD not set, exiting..."
    exit 1
fi

cur_version=""
# mysql only.
DBCNF="-hlocalhost -u${DB_USR}"

# For current migrator, we need to use the old pwd as the password of pgsql.
PGSQL_USR="postgres"
POSTGRES_PASSWORD=${DB_PWD}

file_env() {
        local var="$1"
        local fileVar="${var}_FILE"
        local def="${2:-}"
        if [ "${!var:-}" ] && [ "${!fileVar:-}" ]; then
                echo >&2 "error: both $var and $fileVar are set (but are exclusive)"
                exit 1
        fi
        local val="$def"
        if [ "${!var:-}" ]; then
                val="${!var}"
        elif [ "${!fileVar:-}" ]; then
                val="$(< "${!fileVar}")"
        fi
        export "$var"="$val"
        unset "$fileVar"
}

if [ "${1:0:1}" = '-' ]; then
        set -- postgres "$@"
fi

function get_version {
    set +e
    if [ $ISMYSQL == true ]; then
        launch_mysql
        if [[ $(mysql $DBCNF -N -s -e "select count(*) from information_schema.tables \
            where table_schema='registry' and table_name='alembic_version';") -eq 0 ]]; then
            echo "table alembic_version does not exist. Trying to initial alembic_version."
            mysql $DBCNF < ./alembic.sql
            #compatible with version 0.1.0 and 0.1.1
            if [[ $(mysql $DBCNF -N -s -e "select count(*) from information_schema.tables \
                where table_schema='registry' and table_name='properties'") -eq 0 ]]; then
                echo "table properties does not exist. The version of registry is 0.1.0"
                cur_version='0.1.0'
            else
                echo "The version of registry is 0.1.1"
                mysql $DBCNF -e "insert into registry.alembic_version values ('0.1.1')"
                cur_version='0.1.1'
            fi
        else
            cur_version=$(mysql $DBCNF -N -s -e "select * from registry.alembic_version;")
            echo $cur_version
        fi
    fi
    set -e
    if [ $ISPGSQL == true ]; then
        launch_pgsql $PGSQL_USR
        cur_version=$(psql -U $PGSQL_USR -d registry -t -c "select * from alembic_version;")
        echo $cur_version
    fi
}

function launch_mysql {
    set +e
    local var="$1"
    export MYSQL_PWD="${DB_PWD}"
    echo 'Trying to start mysql server...'
    chown -R 10000:10000 /var/lib/mysql
    mysqld &
    echo 'Waiting for MySQL start...'
    for i in {60..0}; do
        mysqladmin -u$DB_USR -p$DB_PWD processlist >/dev/null 2>&1   
        if [ $? = 0 ]; then
            break
        fi
        sleep 1
    done
    set -e
    if [ "$i" = 0 ]; then
        echo "timeout. Can't run mysql server."
        if [[ $var = "test" ]]; then
            echo "DB test failed."
        fi
        exit 1
    fi
    if [[ $var = "test" ]]; then
        echo "DB test passed."
        exit 0
    fi
}

function stop_mysql {
    if [[ -z $2 ]]; then
        mysqladmin -u$1 shutdown
    else
        mysqladmin -u$1 -p$DB_PWD shutdown
    fi
    sleep 5
}

function version_com() {
    local v1="$1"
    local v2="$2"

    if [ "$v1" = "$v2" ]; then
        return 0
    fi

    ## $v1 is bigger
    if [ "$(echo "$@" | tr " " "\n" | sort -V | head -n 1)" != "$v1" ]; then
        return 1
    fi

    ## $v1 is smaller
    if [ "$(echo "$@" | tr " " "\n" | sort -V | head -n 1)" == "$v1" ]; then
        return 2
    fi

}

function backup {
    echo "Performing backup..."
    if [ $ISMYSQL == true ]; then
        launch_mysql
        mysqldump $DBCNF --add-drop-database --databases registry > /harbor-migration/backup/registry.sql
    fi
    if [ $ISPGSQL == true ]; then
        echo "pg backup"
    fi
    rc="$?"
    echo "Backup performed."
    exit $rc
}

function restore {
    echo "Performing restore..."
    if [ $ISMYSQL == true ]; then
        launch_mysql
        mysql $DBCNF < /harbor-migration/backup/registry.sql
    fi
    if [ $ISPGSQL == true ]; then
        echo "pg restore"
    fi
    rc="$?"
    echo "Restore performed."
    exit $rc
}

function validate {
    if [ $ISMYSQL == true ]; then
        launch_mysql test
    fi
    if [ $ISPGSQL == true ]; then
        launch_pgsql $PGSQL_USR test
    fi
}

function up_harbor {
    local target_version="$1"

    set +e
    version_com $cur_version '1.5.0'
    v1_com=$?
    version_com $target_version '1.5.0'
    v2_com=$?
    ## if no version specific, see it as larger then 1.5.0
    if [ "$target_version" = 'head' ]; then
        v2_com=1
    fi
    set -e

    # $cur_version <='1.5.0', $target_version <='1.5.0', it needs to call mysql upgrade.
    if [ $v1_com != 1 ] && [ $v2_com != 1 ]; then
        if [ $ISMYSQL != true ]; then
            echo "Please make sure to mount the correct the data volumn."
            return 1
        else
            alembic_up $target_version
            return $?
        fi
    fi

    # $cur_version <='1.5.0', $target_version >'1.5.0', it needs to upgrade to 1.5.0.mysql => 1.5.0.pgsql => target_version.pgsql.
    if [ $v1_com != 1 ] && [ $v2_com = 1 ]; then
        if [ $ISMYSQL != true ]; then
            echo "Please make sure to mount the correct the data volumn."
            return 1
        else
            alembic_up '1.5.0' false

            ## dump mysql
            mysqldump --compatible=postgresql --default-character-set=utf8 --databases registry > /harbor-migration/db/registry.mysql     
            stop_mysql $DB_USR $DB_PWD

            ## migrate 1.5.0-mysql to 1.5.0-pqsql.
            python /harbor-migration/db/pgsql_migrator.py /harbor-migration/db/registry.mysql /harbor-migration/db/registry.pgsql

            launch_pgsql $PGSQL_USR
            psql -U $PGSQL_USR -f /harbor-migration/db/schema/registry_from_$cur_version.pgsql
            ##TODO add update notary flag
            psql -U $PGSQL_USR -f /harbor-migration/db/schema/notaryserver.pgsql
            psql -U $PGSQL_USR -f /harbor-migration/db/schema/notarysigner.pgsql
            psql -U $PGSQL_USR -f /harbor-migration/db/registry.pgsql

            ## it needs to call the alembic_up to target, disable it as it's now unsupported.
            #alembic_up $target_version
            stop_pgsql
            return 0
        fi        
    fi

    # $cur_version > '1.5.0', $target_version > '1.5.0', it needs to pgsql upgrade.    
    if [ $v1_com = 1 ] && [ $v2_com = 1 ]; then
        if [ $ISPGSQL != true ]; then
            echo "Please make sure to mount the correct the data volumn."
            return 1
        else
            alembic_up $target_version
            return $?
        fi
    fi

    echo "Unsupported DB upgrade from $cur_version to $target_version, please check the inputs."
    return 1
}

# This function is only for notary <= 1.5.0 mysql to pgsql.
function up_notary {

    set +e
    chown -R 10000:10000 /var/lib/mysql
    mysqld &
    sleep 5
    ls -la /var/lib/mysql

    # for i in {60..0}; do
    #     mysqladmin -uroot processlist >/dev/null 2>&1      
    #     if [ $? = 0 ]; then
    #         break
    #     fi
    #     sleep 1
    # done
    # if [ "$i" = 0 ]; then
    #     echo "timeout. Can't run mysql server."
    #     return 1
    # fi
    set -e

    mysqldump --skip-triggers --compact --no-create-info --skip-quote-names --hex-blob --compatible=postgresql --default-character-set=utf8 --databases notaryserver > /harbor-migration/db/notaryserver.mysql.tmp
    sed "s/0x\([0-9A-F]*\)/decode('\1','hex')/g" /harbor-migration/db/notaryserver.mysql.tmp > /harbor-migration/db/notaryserver.mysql
    mysqldump --skip-triggers --compact --no-create-info --skip-quote-names --hex-blob --compatible=postgresql --default-character-set=utf8 --databases notarysigner > /harbor-migration/db/notarysigner.mysql.tmp    
    sed "s/0x\([0-9A-F]*\)/decode('\1','hex')/g" /harbor-migration/db/notarysigner.mysql.tmp > /harbor-migration/db/notarysigner.mysql
    stop_mysql root

    ## migrate 1.5.0-mysql to 1.5.0-pqsql.
    python /harbor-migration/db/pgsql_migrator.py /harbor-migration/db/notaryserver.mysql /harbor-migration/db/notaryserver.pgsql
    python /harbor-migration/db/pgsql_migrator.py /harbor-migration/db/notarysigner.mysql /harbor-migration/db/notarysigner.pgsql

    ## No signuature, no need to upgrade 
    if grep -q 'INSERT INTO "tuf_files"' "/harbor-migration/db/notaryserver.pgsql"; then
        echo "No need to upgrade notary db as no trust data."
        return 0
    fi 

    # launch_pgsql $PGSQL_USR
    chown -R postgres:postgres $PGDATA
    su - $PGSQL_USR -c "pg_ctl -D \"$PGDATA\" -o \"-c listen_addresses='localhost'\" -w start"
    psql -U $PGSQL_USR -f /harbor-migration/db/notaryserver.pgsql
    psql -U $PGSQL_USR -f /harbor-migration/db/notarysigner.pgsql

    stop_pgsql
    return 0   
 
}

function upgrade {
    up_notary
}

function upgrade1 {

    local target_version="$1"
    if [[ -z $target_version ]]; then
        target_version="head"
        echo "Version is not specified. Default version is head."
    fi

    get_version
    if [ "$cur_version" = "$target_version" ]; then
        echo "It has always running the $target_version, no longer need to upgrade."
        exit 0
    fi

    set +e
    version_com $cur_version '1.5.0'
    v1_com=$?
    set -e

    ls -la /var/lib/mysql

    up_harbor $target_version
    if [ "$?" != 0 ]; then
        echo "Upgrade harbor db error, please run it again."
        exit 1
    fi 
    
    # $cur_version <='1.5.0', it needs to mv data and call notary upgrade
    if [ $v1_com != 1 ]; then
        mkdir -p /harbor-migration/backup
        cp -rf /var/lib/mysql/* /harbor-migration/backup
        rm -rf /var/lib/mysql/*
        if [ "$UPNOTARY" == true ]; then
            ls -la /notary-db
            mv /notary-db/* /var/lib/mysql          
            up_notary
            if [ "$?" != 0 ]; then
                echo "Upgrade notary db error, reverting the data..."
                cp -rf /harbor-migration/backup/* /var/lib/mysql/
                echo "Success to revert data, please run it again..."
                exit 1
            fi 
            rm -rf /var/lib/mysql/*
        fi
        ## move all the data to /data/database
        cp -rf $PGDATA/* /var/lib/mysql
    fi   

}

function alembic_up() {
    local is_exit=true
    if [ "$2" == false ]; then
        is_exit=false
    fi

    export PYTHONPATH=$PYTHONPATH:/harbor-migration/db
    ## TODO: add support for pgsql.
    source /harbor-migration/db/alembic.tpl > /harbor-migration/db/alembic.ini
    
    echo "Performing upgrade $1..."
    alembic -c /harbor-migration/db/alembic.ini current
    alembic -c /harbor-migration/db/alembic.ini upgrade $1
    rc="$?"
    alembic -c /harbor-migration/db/alembic.ini current	
    echo "Upgrade performed."
    if [ $is_exit == true ]; then
        exit $rc
    fi
}

## TODO: add test for pgsql connection.
function launch_pgsql {

    if [ "$1" = 'postgres' ]; then
            chown -R postgres:postgres $PGDATA
            echo here1
            # look specifically for PG_VERSION, as it is expected in the DB dir
            if [ ! -s "$PGDATA/PG_VERSION" ]; then
                    file_env 'POSTGRES_INITDB_ARGS'
                    if [ "$POSTGRES_INITDB_XLOGDIR" ]; then
                            export POSTGRES_INITDB_ARGS="$POSTGRES_INITDB_ARGS --xlogdir $POSTGRES_INITDB_XLOGDIR"
                    fi
                    echo hehe2
                    su - $1 -c "initdb -D $PGDATA  -U postgres -E UTF-8 --lc-collate=en_US.UTF-8 --lc-ctype=en_US.UTF-8 $POSTGRES_INITDB_ARGS"
                    echo hehe3
                    # check password first so we can output the warning before postgres
                    # messes it up
                    file_env 'POSTGRES_PASSWORD'
                    if [ "$POSTGRES_PASSWORD" ]; then
                            pass="PASSWORD '$POSTGRES_PASSWORD'"
                            authMethod=md5
                    else
                            # The - option suppresses leading tabs but *not* spaces. :)
                            echo "Use \"-e POSTGRES_PASSWORD=password\" to set the password in \"docker run\"."
                            exit 1
                    fi

                    {
                            echo
                            echo "host all all all $authMethod"
                    } >> "$PGDATA/pg_hba.conf"
                    # internal start of server in order to allow set-up using psql-client
                    # does not listen on external TCP/IP and waits until start finishes
                    su - $1 -c "pg_ctl -D \"$PGDATA\" -o \"-c listen_addresses='localhost'\" -w start"

                    file_env 'POSTGRES_USER' 'postgres'
                    file_env 'POSTGRES_DB' "$POSTGRES_USER"

                    psql=( psql -v ON_ERROR_STOP=1 )

                    if [ "$POSTGRES_DB" != 'postgres' ]; then
                            "${psql[@]}" --username postgres <<-EOSQL
                                    CREATE DATABASE "$POSTGRES_DB" ;
EOSQL
                            echo
                    fi

                    if [ "$POSTGRES_USER" = 'postgres' ]; then
                            op='ALTER'
                    else
                            op='CREATE'
                    fi
                    "${psql[@]}" --username postgres <<-EOSQL
                            $op USER "$POSTGRES_USER" WITH SUPERUSER $pass ;
EOSQL
                    echo

                    psql+=( --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" )

                    echo
                    for f in /docker-entrypoint-initdb.d/*; do
                            case "$f" in
                                    *.sh)     echo "$0: running $f"; . "$f" ;;
                                    *.sql)    echo "$0: running $f"; "${psql[@]}" -f "$f"; echo ;;
                                    *.sql.gz) echo "$0: running $f"; gunzip -c "$f" | "${psql[@]}"; echo ;;
                                    *)        echo "$0: ignoring $f" ;;
                            esac
                            echo
                    done

                    #PGUSER="${PGUSER:-postgres}" \
                    #su - $1 -c "pg_ctl -D \"$PGDATA\" -m fast -w stop"

                    echo
                    echo 'PostgreSQL init process complete; ready for start up.'
                    echo
            fi
    fi
}

function stop_pgsql {
    su - $PGSQL_USR -c "pg_ctl -D \"/var/lib/postgresql/data\" -w stop"
}

function main {

    if [[ $1 = "help" || $1 = "h" || $# = 0 ]]; then
        echo "Usage:"
        echo "backup                perform database backup"
        echo "restore               perform database restore"
        echo "up,   upgrade         perform database schema upgrade"
        echo "test                  test database connection"
        echo "h,    help            usage help"
        exit 0
    fi

    local key="$1"

    case $key in
    up|upgrade)
        upgrade $2
        ;;    
    backup)
       backup
        ;;
    restore)
       restore
        ;;
    test)
       validate
        ;;
    *)
        echo "unknown option"
        exit 0
        ;;
    esac       
}

main "$@"