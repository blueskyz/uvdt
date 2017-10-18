-- create uvdt database
-- create database uvdt /*!40100 DEFAULT CHARACTER SET utf8mb4 */;

use uvdt

-- create torrent table
create table if not exists `infohash` (
    `infohash` varchar(64) primary key comment '40 bit sha1',
    `name` varchar(128) comment 'file name',
    `peers` varchar(4096) comment 'json [peer_id:ip:port]',
    `ctime` bigint NOT NULL DEFAULT 0,
    `mtime` bigint NOT NULL DEFAULT 0,
    `status` tinyint NOT NULL DEFAULT 0, -- -2 delete, -1 disable, 0 normal
    `torrent` text
) engine=innodb default charset=utf8mb4;
