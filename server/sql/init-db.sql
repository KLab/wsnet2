create database wsnet charset utf8mb4;
create user 'wsnet'@'localhost' identified by 'wsnetpass';
grant all on wsnet.* to 'wsnet'@'localhost';
grant all on wsnet.* to 'wsnet'@'%';

