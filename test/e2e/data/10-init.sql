CREATE DATABASE `sms-public`;
CREATE DATABASE `sms-private`;

CREATE USER 'sms' @'%' IDENTIFIED BY 'sms';
GRANT ALL PRIVILEGES ON `sms-public`.* TO 'sms' @'%';
GRANT ALL PRIVILEGES ON `sms-private`.* TO 'sms' @'%';