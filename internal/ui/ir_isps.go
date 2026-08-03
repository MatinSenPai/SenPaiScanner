package ui

import (
	"bytes"
	"net"
	"strings"
)

type irRange struct {
	from net.IP
	to   net.IP
	isp  string
}

var irRanges []irRange

const rawIRRanges = `
2.57.3.0	2.57.3.255	Enteghal Dadeh Mahan Co. PJSC
2.144.0.0	2.144.0.255	Iran Cell Service and Communication Company
2.144.1.0	2.144.1.255	Iran Cell Service and Communication Company
2.144.2.0	2.144.2.255	Iran Cell Service and Communication Company
2.144.3.0	2.144.3.255	Iran Cell Service and Communication Company
2.144.4.0	2.144.5.255	Iran Cell Service and Communication Company
2.144.6.0	2.144.6.255	Iran Cell Service and Communication Company
2.144.7.0	2.144.7.255	Iran Cell Service and Communication Company
2.144.8.0	2.144.8.255	Iran Cell Service and Communication Company
2.144.9.0	2.144.9.255	Iran Cell Service and Communication Company
2.144.10.0	2.144.11.255	Iran Cell Service and Communication Company
2.144.12.0	2.144.12.255	Iran Cell Service and Communication Company
2.144.13.0	2.144.13.255	Iran Cell Service and Communication Company
2.144.14.0	2.144.15.255	Iran Cell Service and Communication Company
2.144.16.0	2.144.19.255	Iran Cell Service and Communication Company
2.144.20.0	2.144.23.255	Iran Cell Service and Communication Company
2.144.24.0	2.144.31.255	Iran Cell Service and Communication Company
2.144.32.0	2.144.63.255	Iran Cell Service and Communication Company
2.144.64.0	2.144.127.255	Iran Cell Service and Communication Company
2.144.128.0	2.144.191.255	Iran Cell Service and Communication Company
2.144.192.0	2.144.192.255	Iran Cell Service and Communication Company
2.144.193.0	2.144.193.255	Iran Cell Service and Communication Company
2.144.194.0	2.144.195.255	Iran Cell Service and Communication Company
2.144.196.0	2.144.199.255	Iran Cell Service and Communication Company
2.144.200.0	2.144.207.255	Iran Cell Service and Communication Company
2.144.208.0	2.144.223.255	Iran Cell Service and Communication Company
2.144.224.0	2.144.239.255	Iran Cell Service and Communication Company
2.144.240.0	2.144.241.255	Iran Cell Service and Communication Company
2.144.242.0	2.144.243.255	Iran Cell Service and Communication Company
2.144.244.0	2.144.247.255	Iran Cell Service and Communication Company
2.144.248.0	2.144.255.255	Iran Cell Service and Communication Company
2.145.0.0	2.145.15.255	Iran Cell Service and Communication Company
2.145.16.0	2.145.31.255	Iran Cell Service and Communication Company
2.145.32.0	2.145.47.255	Iran Cell Service and Communication Company
2.145.48.0	2.145.63.255	Iran Cell Service and Communication Company
2.145.64.0	2.145.79.255	Iran Cell Service and Communication Company
2.145.80.0	2.145.95.255	Iran Cell Service and Communication Company
2.145.96.0	2.145.111.255	Iran Cell Service and Communication Company
2.145.112.0	2.145.127.255	Iran Cell Service and Communication Company
2.145.128.0	2.145.143.255	Iran Cell Service and Communication Company
2.145.144.0	2.145.159.255	Iran Cell Service and Communication Company
2.145.160.0	2.145.175.255	Iran Cell Service and Communication Company
2.145.176.0	2.145.191.255	Iran Cell Service and Communication Company
2.145.192.0	2.145.207.255	Iran Cell Service and Communication Company
2.145.208.0	2.145.223.255	Iran Cell Service and Communication Company
2.145.224.0	2.145.239.255	Iran Cell Service and Communication Company
2.145.240.0	2.145.255.255	Iran Cell Service and Communication Company
2.146.0.0	2.146.63.255	Iran Cell Service and Communication Company
2.146.64.0	2.146.64.255	Iran Cell Service and Communication Company
2.146.65.0	2.146.65.255	Iran Cell Service and Communication Company
2.146.66.0	2.146.67.255	Iran Cell Service and Communication Company
2.146.68.0	2.146.71.255	Iran Cell Service and Communication Company
2.146.72.0	2.146.79.255	Iran Cell Service and Communication Company
2.146.80.0	2.146.95.255	Iran Cell Service and Communication Company
2.146.96.0	2.146.127.255	Iran Cell Service and Communication Company
2.146.128.0	2.146.255.255	Iran Cell Service and Communication Company
2.147.0.0	2.147.3.255	Iran Cell Service and Communication Company
2.147.4.0	2.147.4.255	Iran Cell Service and Communication Company
2.147.5.0	2.147.5.255	Iran Cell Service and Communication Company
2.147.6.0	2.147.7.255	Iran Cell Service and Communication Company
2.147.8.0	2.147.15.255	Iran Cell Service and Communication Company
2.147.16.0	2.147.23.255	Iran Cell Service and Communication Company
2.147.24.0	2.147.31.255	Iran Cell Service and Communication Company
2.147.32.0	2.147.39.255	Iran Cell Service and Communication Company
2.147.40.0	2.147.47.255	Iran Cell Service and Communication Company
2.147.48.0	2.147.55.255	Iran Cell Service and Communication Company
2.147.56.0	2.147.63.255	Iran Cell Service and Communication Company
2.147.64.0	2.147.71.255	Iran Cell Service and Communication Company
2.147.72.0	2.147.79.255	Iran Cell Service and Communication Company
2.147.80.0	2.147.87.255	Iran Cell Service and Communication Company
2.147.88.0	2.147.95.255	Iran Cell Service and Communication Company
2.147.96.0	2.147.103.255	Iran Cell Service and Communication Company
2.147.104.0	2.147.111.255	Iran Cell Service and Communication Company
2.147.112.0	2.147.119.255	Iran Cell Service and Communication Company
2.147.120.0	2.147.127.255	Iran Cell Service and Communication Company
2.147.128.0	2.147.135.255	Iran Cell Service and Communication Company
2.147.136.0	2.147.143.255	Iran Cell Service and Communication Company
2.147.144.0	2.147.151.255	Iran Cell Service and Communication Company
2.147.152.0	2.147.159.255	Iran Cell Service and Communication Company
2.147.160.0	2.147.167.255	Iran Cell Service and Communication Company
2.147.168.0	2.147.175.255	Iran Cell Service and Communication Company
2.147.176.0	2.147.183.255	Iran Cell Service and Communication Company
2.147.184.0	2.147.191.255	Iran Cell Service and Communication Company
2.147.192.0	2.147.199.255	Iran Cell Service and Communication Company
2.147.200.0	2.147.207.255	Iran Cell Service and Communication Company
2.147.208.0	2.147.215.255	Iran Cell Service and Communication Company
2.147.216.0	2.147.223.255	Iran Cell Service and Communication Company
2.147.224.0	2.147.231.255	Iran Cell Service and Communication Company
2.147.232.0	2.147.239.255	Iran Cell Service and Communication Company
2.147.240.0	2.147.247.255	Iran Cell Service and Communication Company
2.147.248.0	2.147.255.255	Iran Cell Service and Communication Company
2.176.0.0	2.177.255.255	Iran Telecommunication Company Pjs
2.178.0.0	2.178.255.255	Iran Telecommunication Company Pjs
2.179.0.0	2.179.7.255	Iran Telecommunication Company Pjs
2.179.8.0	2.179.11.255	Iran Telecommunication Company Pjs
2.179.12.0	2.179.15.255	Iran Telecommunication Company Pjs
2.179.16.0	2.179.19.255	Iran Telecommunication Company Pjs
2.179.20.0	2.179.23.255	Iran Telecommunication Company Pjs
2.179.24.0	2.179.27.255	Iran Telecommunication Company Pjs
2.179.28.0	2.179.31.255	Iran Telecommunication Company Pjs
2.179.32.0	2.179.39.255	Iran Telecommunication Company Pjs
2.179.40.0	2.179.47.255	Iran Telecommunication Company Pjs
2.179.48.0	2.179.51.255	Iran Telecommunication Company Pjs
2.179.52.0	2.179.53.255	Iran Telecommunication Company Pjs
2.179.54.0	2.179.55.255	Iran Telecommunication Company Pjs
2.179.56.0	2.179.59.255	Iran Information Technology Company PJSC
2.179.60.0	2.179.61.255	Iran Telecommunication Company Pjs
2.179.62.0	2.179.63.255	Iran Telecommunication Company Pjs
2.179.64.0	2.179.79.255	Iran Telecommunication Company Pjs
2.179.80.0	2.179.95.255	Iran Telecommunication Company Pjs
2.179.96.0	2.179.111.255	Iran Telecommunication Company Pjs
2.179.112.0	2.179.127.255	Iran Telecommunication Company Pjs
2.179.128.0	2.179.159.255	Iran Telecommunication Company Pjs
2.179.160.0	2.179.163.255	Iran Telecommunication Company Pjs
2.179.164.0	2.179.167.255	Iran Telecommunication Company Pjs
2.179.168.0	2.179.168.255	Iran Information Technology Company PJSC
2.179.169.0	2.179.169.255	Iran Telecommunication Company Pjs
2.179.170.0	2.179.170.255	Iran Telecommunication Company Pjs
2.179.171.0	2.179.171.255	Iran Telecommunication Company Pjs
2.179.172.0	2.179.175.255	Iran Telecommunication Company Pjs
2.179.176.0	2.179.191.255	Iran Telecommunication Company Pjs
2.179.192.0	2.179.199.255	Iran Telecommunication Company Pjs
2.179.200.0	2.179.207.255	Iran Telecommunication Company Pjs
2.179.208.0	2.179.223.255	Iran Telecommunication Company Pjs
2.179.224.0	2.179.239.255	Iran Telecommunication Company Pjs
2.179.240.0	2.179.255.255	Iran Telecommunication Company Pjs
2.180.0.0	2.180.127.255	Iran Telecommunication Company Pjs
2.180.128.0	2.180.191.255	Iran Telecommunication Company Pjs
2.180.192.0	2.180.223.255	Iran Telecommunication Company Pjs
2.180.224.0	2.180.239.255	Iran Telecommunication Company Pjs
2.180.240.0	2.180.247.255	Iran Telecommunication Company Pjs
2.180.248.0	2.180.251.255	Iran Telecommunication Company Pjs
2.180.252.0	2.180.253.255	Iran Telecommunication Company Pjs
2.180.254.0	2.180.254.255	Iran Information Technology Company PJSC
2.180.255.0	2.180.255.255	Iran Telecommunication Company Pjs
2.181.0.0	2.181.31.255	Iran Telecommunication Company Pjs
2.181.32.0	2.181.35.255	Iran Telecommunication Company Pjs
2.181.36.0	2.181.39.255	Iran Telecommunication Company Pjs
2.181.40.0	2.181.47.255	Iran Telecommunication Company Pjs
2.181.48.0	2.181.55.255	Iran Telecommunication Company Pjs
2.181.56.0	2.181.63.255	Iran Telecommunication Company Pjs
2.181.64.0	2.181.79.255	Iran Telecommunication Company Pjs
2.181.80.0	2.181.95.255	Iran Telecommunication Company Pjs
2.181.96.0	2.181.111.255	Iran Telecommunication Company Pjs
2.181.112.0	2.181.127.255	Iran Telecommunication Company Pjs
2.181.128.0	2.181.159.255	Iran Telecommunication Company Pjs
2.181.160.0	2.181.175.255	Iran Telecommunication Company Pjs
2.181.176.0	2.181.177.255	Iran Telecommunication Company Pjs
2.181.178.0	2.181.178.255	Iran Telecommunication Company Pjs
2.181.179.0	2.181.179.255	Iran Telecommunication Company Pjs
2.181.180.0	2.181.183.255	Iran Telecommunication Company Pjs
2.181.184.0	2.181.187.255	Iran Information Technology Company PJSC
2.181.188.0	2.181.191.255	Iran Telecommunication Company Pjs
2.181.192.0	2.181.193.255	Iran Information Technology Company PJSC
2.181.194.0	2.181.195.255	Iran Telecommunication Company Pjs
2.181.196.0	2.181.199.255	Iran Telecommunication Company Pjs
2.181.200.0	2.181.207.255	Iran Telecommunication Company Pjs
2.181.208.0	2.181.215.255	Iran Telecommunication Company Pjs
2.181.216.0	2.181.219.255	Iran Telecommunication Company Pjs
2.181.220.0	2.181.221.255	Iran Telecommunication Company Pjs
2.181.222.0	2.181.222.255	Iran Information Technology Company PJSC
2.181.223.0	2.181.223.255	Iran Information Technology Company PJSC
2.181.224.0	2.181.255.255	Iran Telecommunication Company Pjs
2.182.0.0	2.182.31.255	Iran Telecommunication Company Pjs
2.182.32.0	2.182.39.255	Iran Telecommunication Company Pjs
2.182.40.0	2.182.43.255	Iran Telecommunication Company Pjs
2.182.44.0	2.182.47.255	Iran Telecommunication Company Pjs
2.182.48.0	2.182.63.255	Iran Telecommunication Company Pjs
2.182.64.0	2.182.95.255	Iran Telecommunication Company Pjs
2.182.96.0	2.182.97.255	Iran Telecommunication Company Pjs
2.182.98.0	2.182.99.255	Iran Telecommunication Company Pjs
2.182.100.0	2.182.103.255	Iran Telecommunication Company Pjs
2.182.104.0	2.182.107.255	Iran Telecommunication Company Pjs
2.182.108.0	2.182.109.255	Iran Telecommunication Company Pjs
2.182.110.0	2.182.111.255	Iran Telecommunication Company Pjs
2.182.112.0	2.182.119.255	Iran Telecommunication Company Pjs
2.182.120.0	2.182.123.255	Iran Telecommunication Company Pjs
2.182.124.0	2.182.127.255	Iran Telecommunication Company Pjs
2.182.128.0	2.182.135.255	Iran Telecommunication Company Pjs
2.182.136.0	2.182.143.255	Iran Information Technology Company PJSC
2.182.144.0	2.182.151.255	Iran Telecommunication Company Pjs
2.182.152.0	2.182.155.255	Iran Telecommunication Company Pjs
2.182.156.0	2.182.159.255	Iran Telecommunication Company Pjs
2.182.160.0	2.182.161.255	Iran Telecommunication Company Pjs
2.182.162.0	2.182.163.255	Iran Telecommunication Company Pjs
2.182.164.0	2.182.167.255	Iran Telecommunication Company Pjs
2.182.168.0	2.182.169.255	Iran Telecommunication Company Pjs
2.182.170.0	2.182.171.255	Iran Telecommunication Company Pjs
2.182.172.0	2.182.172.255	Iran Information Technology Company PJSC
2.182.173.0	2.182.173.255	Iran Telecommunication Company Pjs
2.182.174.0	2.182.174.255	Iran Information Technology Company PJSC
2.182.175.0	2.182.175.255	Iran Telecommunication Company Pjs
2.182.176.0	2.182.191.255	Iran Telecommunication Company Pjs
2.182.192.0	2.182.199.255	Iran Telecommunication Company Pjs
2.182.200.0	2.182.200.255	Iran Telecommunication Company Pjs
2.182.201.0	2.182.201.255	Iran Telecommunication Company Pjs
2.182.202.0	2.182.203.255	Iran Telecommunication Company Pjs
2.182.204.0	2.182.207.255	Iran Telecommunication Company Pjs
2.182.208.0	2.182.211.255	Iran Telecommunication Company Pjs
2.182.212.0	2.182.213.255	Iran Telecommunication Company Pjs
2.182.214.0	2.182.215.255	Iran Telecommunication Company Pjs
2.182.216.0	2.182.223.255	Iran Telecommunication Company Pjs
2.182.224.0	2.182.239.255	Iran Telecommunication Company Pjs
2.182.240.0	2.182.247.255	Iran Telecommunication Company Pjs
2.182.248.0	2.182.251.255	Iran Telecommunication Company Pjs
2.182.252.0	2.182.252.255	Iran Telecommunication Company Pjs
2.182.253.0	2.182.253.255	Iran Telecommunication Company Pjs
2.182.254.0	2.182.254.255	Iran Telecommunication Company Pjs
2.182.255.0	2.182.255.255	Iran Information Technology Company PJSC
2.183.0.0	2.183.15.255	Iran Telecommunication Company Pjs
2.183.16.0	2.183.31.255	Iran Telecommunication Company Pjs
2.183.32.0	2.183.47.255	Iran Telecommunication Company Pjs
2.183.48.0	2.183.55.255	Iran Telecommunication Company Pjs
2.183.56.0	2.183.59.255	Iran Telecommunication Company Pjs
2.183.60.0	2.183.63.255	Iran Telecommunication Company Pjs
2.183.64.0	2.183.79.255	Iran Telecommunication Company Pjs
2.183.80.0	2.183.95.255	Iran Telecommunication Company Pjs
2.183.96.0	2.183.111.255	Iran Telecommunication Company Pjs
2.183.112.0	2.183.119.255	Iran Telecommunication Company Pjs
2.183.120.0	2.183.123.255	Iran Telecommunication Company Pjs
2.183.124.0	2.183.127.255	Iran Telecommunication Company Pjs
2.183.128.0	2.183.131.255	Iran Telecommunication Company Pjs
2.183.132.0	2.183.135.255	Iran Telecommunication Company Pjs
2.183.136.0	2.183.139.255	Iran Telecommunication Company Pjs
2.183.140.0	2.183.141.255	Iran Telecommunication Company Pjs
2.183.142.0	2.183.143.255	Iran Telecommunication Company Pjs
2.183.144.0	2.183.151.255	Iran Telecommunication Company Pjs
2.183.152.0	2.183.155.255	Iran Telecommunication Company Pjs
2.183.156.0	2.183.159.255	Iran Telecommunication Company Pjs
2.183.160.0	2.183.175.255	Iran Telecommunication Company Pjs
2.183.176.0	2.183.183.255	Iran Telecommunication Company Pjs
2.183.184.0	2.183.187.255	Iran Telecommunication Company Pjs
2.183.188.0	2.183.191.255	Iran Telecommunication Company Pjs
2.183.192.0	2.183.199.255	Iran Telecommunication Company Pjs
2.183.200.0	2.183.201.255	Iran Telecommunication Company Pjs
2.183.202.0	2.183.203.255	Iran Telecommunication Company Pjs
2.183.204.0	2.183.207.255	Iran Telecommunication Company Pjs
2.183.208.0	2.183.223.255	Iran Telecommunication Company Pjs
2.183.224.0	2.183.231.255	Iran Telecommunication Company Pjs
2.183.232.0	2.183.233.255	Iran Telecommunication Company Pjs
2.183.234.0	2.183.235.255	Iran Telecommunication Company Pjs
2.183.236.0	2.183.237.255	Iran Telecommunication Company Pjs
2.183.238.0	2.183.239.255	Iran Telecommunication Company Pjs
2.183.240.0	2.183.255.255	Iran Telecommunication Company Pjs
2.184.0.0	2.184.31.255	Iran Telecommunication Company Pjs
2.184.32.0	2.184.63.255	Iran Telecommunication Company Pjs
2.184.64.0	2.184.67.255	Iran Telecommunication Company Pjs
2.184.68.0	2.184.69.255	Iran Telecommunication Company Pjs
2.184.70.0	2.184.71.255	Iran Telecommunication Company Pjs
2.184.72.0	2.184.79.255	Iran Telecommunication Company Pjs
2.184.80.0	2.184.95.255	Iran Telecommunication Company Pjs
2.184.96.0	2.184.127.255	Iran Telecommunication Company Pjs
2.184.128.0	2.184.159.255	Iran Telecommunication Company Pjs
2.184.160.0	2.184.161.255	Iran Telecommunication Company Pjs
2.184.162.0	2.184.163.255	Iran Telecommunication Company Pjs
2.184.164.0	2.184.167.255	Iran Telecommunication Company Pjs
2.184.168.0	2.184.169.255	Iran Telecommunication Company Pjs
2.184.170.0	2.184.171.255	Iran Telecommunication Company Pjs
2.184.172.0	2.184.175.255	Iran Telecommunication Company Pjs
2.184.176.0	2.184.191.255	Iran Telecommunication Company Pjs
2.184.192.0	2.184.255.255	Iran Telecommunication Company Pjs
2.185.0.0	2.185.63.255	Iran Telecommunication Company Pjs
2.185.64.0	2.185.95.255	Iran Telecommunication Company Pjs
2.185.96.0	2.185.127.255	Iran Telecommunication Company Pjs
2.185.128.0	2.185.159.255	Iran Telecommunication Company Pjs
2.185.160.0	2.185.191.255	Iran Telecommunication Company Pjs
2.185.192.0	2.185.223.255	Iran Telecommunication Company Pjs
2.185.224.0	2.185.239.255	Iran Telecommunication Company Pjs
2.185.240.0	2.185.247.255	Iran Telecommunication Company Pjs
2.185.248.0	2.185.251.255	Iran Telecommunication Company Pjs
2.185.252.0	2.185.253.255	Iran Telecommunication Company Pjs
2.185.254.0	2.185.254.255	Iran Information Technology Company PJSC
2.185.255.0	2.185.255.255	Iran Information Technology Company PJSC
2.186.0.0	2.186.255.255	Iran Telecommunication Company Pjs
2.187.0.0	2.187.3.255	Iran Telecommunication Company Pjs
2.187.4.0	2.187.5.255	Iran Telecommunication Company Pjs
2.187.6.0	2.187.7.255	Iran Telecommunication Company Pjs
2.187.8.0	2.187.11.255	Iran Telecommunication Company Pjs
2.187.12.0	2.187.13.255	Iran Telecommunication Company Pjs
2.187.14.0	2.187.14.255	Iran Telecommunication Company Pjs
2.187.15.0	2.187.15.255	Iran Telecommunication Company Pjs
2.187.16.0	2.187.16.255	Iran Telecommunication Company Pjs
2.187.17.0	2.187.17.255	Iran Telecommunication Company Pjs
2.187.18.0	2.187.18.255	Iran Telecommunication Company Pjs
2.187.19.0	2.187.19.255	Iran Telecommunication Company Pjs
2.187.20.0	2.187.23.255	Iran Telecommunication Company Pjs
2.187.24.0	2.187.31.255	Iran Telecommunication Company Pjs
2.187.32.0	2.187.47.255	Iran Telecommunication Company Pjs
2.187.48.0	2.187.55.255	Iran Telecommunication Company Pjs
2.187.56.0	2.187.63.255	Iran Telecommunication Company Pjs
2.187.64.0	2.187.95.255	Iran Information Technology Company PJSC
2.187.96.0	2.187.127.255	Iran Telecommunication Company Pjs
2.187.128.0	2.187.159.255	Iran Telecommunication Company Pjs
2.187.160.0	2.187.191.255	Iran Telecommunication Company Pjs
2.187.192.0	2.187.207.255	Iran Telecommunication Company Pjs
2.187.208.0	2.187.223.255	Iran Telecommunication Company Pjs
2.187.224.0	2.187.239.255	Iran Telecommunication Company Pjs
2.187.240.0	2.187.247.255	Iran Telecommunication Company Pjs
2.187.248.0	2.187.251.255	Iran Telecommunication Company Pjs
2.187.252.0	2.187.252.255	Iran Information Technology Company PJSC
2.187.253.0	2.187.253.255	Iran Information Technology Company PJSC
2.187.254.0	2.187.254.255	Iran Information Technology Company PJSC
2.187.255.0	2.187.255.255	Iran Information Technology Company PJSC
2.188.0.0	2.188.0.255	Telecommunication Infrastructure Company
2.188.1.0	2.188.1.255	Telecommunication Infrastructure Company
2.188.2.0	2.188.2.255	Iran Information Technology Company PJSC
2.188.3.0	2.188.3.255	Iran Information Technology Company PJSC
2.188.4.0	2.188.4.255	Iran Information Technology Company PJSC
2.188.5.0	2.188.5.255	Telecommunication Infrastructure Company
2.188.6.0	2.188.6.255	Iran Information Technology Company PJSC
2.188.7.0	2.188.7.255	Telecommunication Infrastructure Company
2.188.8.0	2.188.8.255	Iran Information Technology Company PJSC
2.188.9.0	2.188.9.255	Telecommunication Infrastructure Company
2.188.10.0	2.188.10.255	Iran Information Technology Company PJSC
2.188.11.0	2.188.11.255	Telecommunication Infrastructure Company
2.188.12.0	2.188.12.255	Iran Information Technology Company PJSC
2.188.13.0	2.188.13.255	Iran Information Technology Company PJSC
2.188.14.0	2.188.14.255	Iran Information Technology Company PJSC
2.188.15.0	2.188.15.255	Iran Information Technology Company PJSC
2.188.16.0	2.188.17.255	Telecommunication Infrastructure Company
2.188.18.0	2.188.18.255	Iran Information Technology Company PJSC
2.188.19.0	2.188.19.255	Iran Information Technology Company PJSC
2.188.20.0	2.188.20.255	Iran Information Technology Company PJSC
2.188.21.0	2.188.21.255	Telecommunication Infrastructure Company
2.188.22.0	2.188.22.255	Telecommunication Infrastructure Company
2.188.23.0	2.188.23.255	Telecommunication Infrastructure Company
2.188.24.0	2.188.24.255	Iran Information Technology Company PJSC
2.188.25.0	2.188.25.255	Iran Information Technology Company PJSC
2.188.26.0	2.188.27.255	Iran Information Technology Company PJSC
2.188.28.0	2.188.28.255	Iran Information Technology Company PJSC
2.188.29.0	2.188.29.255	Iran Information Technology Company PJSC
2.188.30.0	2.188.30.255	Telecommunication Infrastructure Company
2.188.31.0	2.188.31.255	Iran Information Technology Company PJSC
2.188.32.0	2.188.39.255	Mobin Net Communication Company (Private Joint Stock)
2.188.40.0	2.188.40.255	Respina Networks & Beyond PJSC
2.188.41.0	2.188.41.255	Respina Networks & Beyond PJSC
2.188.42.0	2.188.42.255	Respina Networks & Beyond PJSC
2.188.43.0	2.188.43.255	Respina Networks & Beyond PJSC
2.188.44.0	2.188.47.255	Telecommunication Infrastructure Company
2.188.48.0	2.188.51.255	Iran Information Technology Company PJSC
2.188.52.0	2.188.52.255	Iran Information Technology Company PJSC
2.188.53.0	2.188.53.255	Iran Information Technology Company PJSC
2.188.54.0	2.188.54.255	Iran Information Technology Company PJSC
2.188.55.0	2.188.55.255	Iran Information Technology Company PJSC
2.188.56.0	2.188.56.255	Iran Information Technology Company PJSC
2.188.57.0	2.188.57.255	Iran Information Technology Company PJSC
2.188.58.0	2.188.58.255	Iran Information Technology Company PJSC
2.188.59.0	2.188.59.255	Iran Information Technology Company PJSC
2.188.60.0	2.188.60.255	Iran Information Technology Company PJSC
2.188.61.0	2.188.61.255	Iran Information Technology Company PJSC
2.188.62.0	2.188.63.255	Telecommunication Infrastructure Company
2.188.64.0	2.188.71.255	Iran Information Technology Company PJSC
2.188.72.0	2.188.72.255	Respina Networks & Beyond PJSC
2.188.73.0	2.188.73.255	Respina Networks & Beyond PJSC
2.188.74.0	2.188.74.255	Respina Networks & Beyond PJSC
2.188.75.0	2.188.75.255	Respina Networks & Beyond PJSC
2.188.76.0	2.188.76.255	Telecommunication Infrastructure Company
2.188.77.0	2.188.77.255	Iran Information Technology Company PJSC
2.188.78.0	2.188.78.255	Iran Information Technology Company PJSC
2.188.79.0	2.188.79.255	Iran Information Technology Company PJSC
2.188.80.0	2.188.95.255	Telecommunication Infrastructure Company
2.188.96.0	2.188.127.255	Telecommunication Infrastructure Company
2.188.128.0	2.188.159.255	Telecommunication Infrastructure Company
2.188.160.0	2.188.160.255	Respina Networks & Beyond PJSC
2.188.161.0	2.188.161.255	Respina Networks & Beyond PJSC
2.188.162.0	2.188.162.255	Respina Networks & Beyond PJSC
2.188.163.0	2.188.163.255	Respina Networks & Beyond PJSC
2.188.164.0	2.188.164.255	Respina Networks & Beyond PJSC
2.188.165.0	2.188.165.255	Respina Networks & Beyond PJSC
2.188.166.0	2.188.166.255	Respina Networks & Beyond PJSC
2.188.167.0	2.188.167.255	Respina Networks & Beyond PJSC
2.188.168.0	2.188.169.255	Tarahan Shabake Sharif LTD
2.188.170.0	2.188.170.255	Respina Networks & Beyond PJSC
2.188.171.0	2.188.171.255	Tarahan Shabake Sharif LTD
2.188.172.0	2.188.175.255	Tarahan Shabake Sharif LTD
2.188.176.0	2.188.177.255	Telecommunication Infrastructure Company
2.188.178.0	2.188.178.255	Iran Information Technology Company PJSC
2.188.179.0	2.188.179.255	Telecommunication Infrastructure Company
2.188.180.0	2.188.180.255	Telecommunication Infrastructure Company
2.188.181.0	2.188.181.255	Iran Information Technology Company PJSC
2.188.182.0	2.188.183.255	Telecommunication Infrastructure Company
2.188.184.0	2.188.184.255	Telecommunication Infrastructure Company
2.188.185.0	2.188.185.255	Telecommunication Infrastructure Company
2.188.186.0	2.188.186.255	Iran Information Technology Company PJSC
2.188.187.0	2.188.187.255	Telecommunication Infrastructure Company
2.188.188.0	2.188.189.255	Telecommunication Infrastructure Company
2.188.190.0	2.188.190.255	Telecommunication Infrastructure Company
2.188.191.0	2.188.191.255	Iran Information Technology Company PJSC
2.188.192.0	2.188.207.255	Telecommunication Infrastructure Company
2.188.208.0	2.188.208.255	Respina Networks & Beyond PJSC
2.188.209.0	2.188.209.255	Respina Networks & Beyond PJSC
2.188.210.0	2.188.210.255	Respina Networks & Beyond PJSC
2.188.211.0	2.188.211.255	Respina Networks & Beyond PJSC
2.188.212.0	2.188.212.255	Respina Networks & Beyond PJSC
2.188.213.0	2.188.213.255	Respina Networks & Beyond PJSC
2.188.214.0	2.188.215.255	Respina Networks & Beyond PJSC
2.188.216.0	2.188.216.255	Respina Networks & Beyond PJSC
2.188.217.0	2.188.217.255	Respina Networks & Beyond PJSC
2.188.218.0	2.188.218.255	Respina Networks & Beyond PJSC
2.188.219.0	2.188.219.255	Respina Networks & Beyond PJSC
2.188.220.0	2.188.220.255	Respina Networks & Beyond PJSC
2.188.221.0	2.188.221.255	Respina Networks & Beyond PJSC
2.188.222.0	2.188.222.255	Respina Networks & Beyond PJSC
2.188.223.0	2.188.223.255	Respina Networks & Beyond PJSC
2.188.224.0	2.188.224.255	Respina Networks & Beyond PJSC
2.188.225.0	2.188.225.255	Respina Networks & Beyond PJSC
2.188.226.0	2.188.226.255	Respina Networks & Beyond PJSC
2.188.227.0	2.188.227.255	Respina Networks & Beyond PJSC
2.188.228.0	2.188.228.255	Respina Networks & Beyond PJSC
2.188.229.0	2.188.229.255	Ofogh Koroush Chain Store Public Joint Stock
2.188.230.0	2.188.230.255	Respina Networks & Beyond PJSC
2.188.231.0	2.188.231.255	Respina Networks & Beyond PJSC
2.188.232.0	2.188.232.255	Respina Networks & Beyond PJSC
2.188.233.0	2.188.233.255	Respina Networks & Beyond PJSC
2.188.234.0	2.188.234.255	Fara Net Gostar Kavoshgaran Aftab
2.188.235.0	2.188.235.255	Respina Networks & Beyond PJSC
2.188.236.0	2.188.236.255	Respina Networks & Beyond PJSC
2.188.237.0	2.188.237.255	Respina Networks & Beyond PJSC
2.188.238.0	2.188.238.255	Fara Net Gostar Kavoshgaran Aftab
2.188.239.0	2.188.239.255	Respina Networks & Beyond PJSC
2.188.240.0	2.188.240.255	Respina Networks & Beyond PJSC
2.188.241.0	2.188.241.255	Respina Networks & Beyond PJSC
2.188.242.0	2.188.242.255	Respina Networks & Beyond PJSC
2.188.243.0	2.188.243.255	Respina Networks & Beyond PJSC
2.188.244.0	2.188.244.255	Respina Networks & Beyond PJSC
2.188.245.0	2.188.245.255	Telecommunication Infrastructure Company
2.188.246.0	2.188.246.255	Respina Networks & Beyond PJSC
2.188.247.0	2.188.247.255	Respina Networks & Beyond PJSC
2.188.248.0	2.188.248.255	Telecommunication Infrastructure Company
2.188.249.0	2.188.249.255	Respina Networks & Beyond PJSC
2.188.250.0	2.188.250.255	Respina Networks & Beyond PJSC
2.188.251.0	2.188.251.255	Respina Networks & Beyond PJSC
2.188.252.0	2.188.252.255	Respina Networks & Beyond PJSC
2.188.253.0	2.188.253.255	Telecommunication Infrastructure Company
2.188.254.0	2.188.255.255	Telecommunication Infrastructure Company
2.189.0.0	2.189.0.255	Iran Information Technology Company PJSC
2.189.1.0	2.189.1.255	Telecommunication Infrastructure Company
2.189.2.0	2.189.2.255	Telecommunication Infrastructure Company
2.189.3.0	2.189.3.255	Telecommunication Infrastructure Company
2.189.4.0	2.189.4.255	Iran Information Technology Company PJSC
2.189.5.0	2.189.5.255	Iran Information Technology Company PJSC
2.189.6.0	2.189.6.255	Telecommunication Infrastructure Company
2.189.7.0	2.189.7.255	Iran Information Technology Company PJSC
2.189.8.0	2.189.11.255	Iran Telecommunication Company Pjs
2.189.12.0	2.189.15.255	Iran Telecommunication Company Pjs
2.189.16.0	2.189.19.255	Iran Telecommunication Company Pjs
2.189.20.0	2.189.23.255	Iran Telecommunication Company Pjs
2.189.24.0	2.189.25.255	Iran Telecommunication Company Pjs
2.189.26.0	2.189.27.255	Iran Information Technology Company PJSC
2.189.28.0	2.189.31.255	Iran Telecommunication Company Pjs
2.189.32.0	2.189.35.255	Iran Telecommunication Company Pjs
2.189.36.0	2.189.37.255	Iran Telecommunication Company Pjs
2.189.38.0	2.189.39.255	Iran Telecommunication Company Pjs
2.189.40.0	2.189.41.255	Telecommunication Infrastructure Company
2.189.42.0	2.189.42.255	Telecommunication Infrastructure Company
2.189.43.0	2.189.43.255	Telecommunication Infrastructure Company
2.189.44.0	2.189.44.255	Telecommunication Infrastructure Company
2.189.45.0	2.189.45.255	Iran Information Technology Company PJSC
2.189.46.0	2.189.47.255	Telecommunication Infrastructure Company
2.189.48.0	2.189.51.255	Telecommunication Infrastructure Company
2.189.52.0	2.189.53.255	Telecommunication Infrastructure Company
2.189.54.0	2.189.55.255	Telecommunication Infrastructure Company
2.189.56.0	2.189.56.255	Andisheh Yavaran Tamadon Emrooz LLC
2.189.57.0	2.189.57.255	Telecommunication Infrastructure Company
2.189.58.0	2.189.58.255	Andisheh Yavaran Tamadon Emrooz LLC
2.189.59.0	2.189.59.255	Iran Information Technology Company PJSC
2.189.60.0	2.189.60.255	Telecommunication Infrastructure Company
2.189.61.0	2.189.61.255	Iran Information Technology Company PJSC
2.189.62.0	2.189.62.255	Iran Information Technology Company PJSC
2.189.63.0	2.189.63.255	Iran Information Technology Company PJSC
2.189.64.0	2.189.64.255	Respina Networks & Beyond PJSC
2.189.65.0	2.189.65.255	Respina Networks & Beyond PJSC
2.189.66.0	2.189.66.255	Respina Networks & Beyond PJSC
2.189.67.0	2.189.67.255	Respina Networks & Beyond PJSC
2.189.68.0	2.189.68.255	Telecommunication Infrastructure Company
2.189.69.0	2.189.69.255	Telecommunication Infrastructure Company
2.189.70.0	2.189.71.255	Telecommunication Infrastructure Company
2.189.72.0	2.189.73.255	Telecommunication Infrastructure Company
2.189.74.0	2.189.74.255	Telecommunication Infrastructure Company
2.189.75.0	2.189.75.255	Iran Information Technology Company PJSC
2.189.76.0	2.189.76.255	Iran Information Technology Company PJSC
2.189.77.0	2.189.77.255	Iran Information Technology Company PJSC
2.189.78.0	2.189.78.255	Iran Information Technology Company PJSC
2.189.79.0	2.189.79.255	Iran Information Technology Company PJSC
2.189.80.0	2.189.81.255	Sharif University of Technology
2.189.82.0	2.189.82.255	Sharif University of Technology
2.189.83.0	2.189.83.255	Sharif University of Technology
2.189.84.0	2.189.84.255	Telecommunication Infrastructure Company
2.189.85.0	2.189.85.255	Respina Networks & Beyond PJSC
2.189.86.0	2.189.86.255	Respina Networks & Beyond PJSC
2.189.87.0	2.189.87.255	Respina Networks & Beyond PJSC
2.189.88.0	2.189.88.255	Respina Networks & Beyond PJSC
2.189.89.0	2.189.89.255	Respina Networks & Beyond PJSC
2.189.90.0	2.189.90.255	Respina Networks & Beyond PJSC
2.189.91.0	2.189.91.255	Respina Networks & Beyond PJSC
2.189.92.0	2.189.92.255	Respina Networks & Beyond PJSC
2.189.93.0	2.189.93.255	Respina Networks & Beyond PJSC
2.189.94.0	2.189.95.255	Respina Networks & Beyond PJSC
2.189.96.0	2.189.96.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
2.189.97.0	2.189.97.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
2.189.98.0	2.189.98.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
2.189.99.0	2.189.99.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
2.189.100.0	2.189.100.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
2.189.101.0	2.189.101.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
2.189.102.0	2.189.102.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
2.189.103.0	2.189.103.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
2.189.104.0	2.189.104.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
2.189.105.0	2.189.105.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
2.189.106.0	2.189.106.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
2.189.107.0	2.189.107.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
2.189.108.0	2.189.108.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
2.189.109.0	2.189.109.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
2.189.110.0	2.189.110.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
2.189.111.0	2.189.111.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
2.189.112.0	2.189.112.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
2.189.113.0	2.189.113.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
2.189.114.0	2.189.114.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
2.189.115.0	2.189.115.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
2.189.116.0	2.189.116.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
2.189.117.0	2.189.117.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
2.189.118.0	2.189.118.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
2.189.119.0	2.189.119.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
2.189.120.0	2.189.120.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
2.189.121.0	2.189.121.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
2.189.122.0	2.189.122.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
2.189.123.0	2.189.123.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
2.189.124.0	2.189.124.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
2.189.125.0	2.189.125.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
2.189.126.0	2.189.126.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
2.189.127.0	2.189.127.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
2.189.128.0	2.189.128.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
2.189.129.0	2.189.129.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
2.189.130.0	2.189.130.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
2.189.131.0	2.189.131.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
2.189.132.0	2.189.132.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
2.189.133.0	2.189.133.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
2.189.134.0	2.189.134.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
2.189.135.0	2.189.135.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
2.189.136.0	2.189.136.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
2.189.137.0	2.189.137.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
2.189.138.0	2.189.138.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
2.189.139.0	2.189.139.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
2.189.140.0	2.189.140.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
2.189.141.0	2.189.141.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
2.189.142.0	2.189.142.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
2.189.143.0	2.189.143.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
2.189.144.0	2.189.144.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
2.189.145.0	2.189.145.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
2.189.146.0	2.189.146.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
2.189.147.0	2.189.147.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
2.189.148.0	2.189.148.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
2.189.149.0	2.189.149.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
2.189.150.0	2.189.150.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
2.189.151.0	2.189.151.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
2.189.152.0	2.189.152.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
2.189.153.0	2.189.153.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
2.189.154.0	2.189.154.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
2.189.155.0	2.189.155.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
2.189.156.0	2.189.156.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
2.189.157.0	2.189.157.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
2.189.158.0	2.189.158.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
2.189.159.0	2.189.159.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
2.189.160.0	2.189.160.255	Respina Networks & Beyond PJSC
2.189.161.0	2.189.161.255	Respina Networks & Beyond PJSC
2.189.162.0	2.189.162.255	Respina Networks & Beyond PJSC
2.189.163.0	2.189.163.255	Respina Networks & Beyond PJSC
2.189.164.0	2.189.164.255	Respina Networks & Beyond PJSC
2.189.165.0	2.189.165.255	Respina Networks & Beyond PJSC
2.189.166.0	2.189.166.255	Respina Networks & Beyond PJSC
2.189.167.0	2.189.167.255	Respina Networks & Beyond PJSC
2.189.168.0	2.189.169.255	Telecommunication Infrastructure Company
2.189.170.0	2.189.170.255	Respina Networks & Beyond PJSC
2.189.171.0	2.189.171.255	Respina Networks & Beyond PJSC
2.189.172.0	2.189.172.255	Respina Networks & Beyond PJSC
2.189.173.0	2.189.173.255	Respina Networks & Beyond PJSC
2.189.174.0	2.189.174.255	Respina Networks & Beyond PJSC
2.189.175.0	2.189.175.255	Respina Networks & Beyond PJSC
2.189.176.0	2.189.183.255	Telecommunication Infrastructure Company
2.189.184.0	2.189.185.255	Telecommunication Infrastructure Company
2.189.186.0	2.189.186.255	Iran Information Technology Company PJSC
2.189.187.0	2.189.187.255	Telecommunication Infrastructure Company
2.189.188.0	2.189.188.255	Telecommunication Infrastructure Company
2.189.189.0	2.189.189.255	Iran Information Technology Company PJSC
2.189.190.0	2.189.191.255	Telecommunication Infrastructure Company
2.189.192.0	2.189.207.255	Telecommunication Infrastructure Company
2.189.208.0	2.189.215.255	Telecommunication Infrastructure Company
2.189.216.0	2.189.219.255	Iran Telecommunication Company Pjs
2.189.220.0	2.189.223.255	Iran Telecommunication Company Pjs
2.189.224.0	2.189.227.255	Iran Telecommunication Company Pjs
2.189.228.0	2.189.231.255	Telecommunication Infrastructure Company
2.189.232.0	2.189.235.255	Iran Telecommunication Company Pjs
2.189.236.0	2.189.239.255	Iran Telecommunication Company Pjs
2.189.240.0	2.189.241.255	Iran Telecommunication Company Pjs
2.189.242.0	2.189.242.255	Asre Pardazeshe Ettelaate Amin Institute
2.189.243.0	2.189.243.255	Asre Pardazeshe Ettelaate Amin Institute
2.189.244.0	2.189.245.255	Iran Telecommunication Company Pjs
2.189.246.0	2.189.247.255	Iran Telecommunication Company Pjs
2.189.248.0	2.189.249.255	Iran Telecommunication Company Pjs
2.189.250.0	2.189.251.255	Iran Telecommunication Company Pjs
2.189.252.0	2.189.253.255	Iran Telecommunication Company Pjs
2.189.254.0	2.189.254.255	Asre Pardazeshe Ettelaate Amin Institute
2.189.255.0	2.189.255.255	Asre Pardazeshe Ettelaate Amin Institute
2.190.0.0	2.191.255.255	Iran Telecommunication Company Pjs
5.1.43.0	5.1.43.255	Sefroyek Pardaz Engineering PJSC
5.10.248.0	5.10.248.255	Saeid Shahrokhi
5.10.249.0	5.10.249.255	Saeid Shahrokhi
5.22.0.0	5.22.15.255	Mobile Communication Company of Iran Plc
5.22.16.0	5.22.31.255	Mobile Communication Company of Iran Plc
5.22.32.0	5.22.47.255	Mobile Communication Company of Iran Plc
5.22.48.0	5.22.63.255	Mobile Communication Company of Iran Plc
5.22.64.0	5.22.79.255	Mobile Communication Company of Iran Plc
5.22.80.0	5.22.95.255	Mobile Communication Company of Iran Plc
5.22.96.0	5.22.111.255	Mobile Communication Company of Iran Plc
5.22.112.0	5.22.127.255	Mobile Communication Company of Iran Plc
5.22.192.0	5.22.199.255	Dadeh Pardazan Sabz Alborz Co.(P.J.S.)
5.22.200.0	5.22.200.255	Peimayesh Ertebat Oxin Co.(PJSC)
5.22.201.0	5.22.201.255	Peimayesh Ertebat Oxin Co.(PJSC)
5.22.202.0	5.22.202.255	Foojan Cloud Computing Co. PJS
5.22.203.0	5.22.203.255	Foojan Cloud Computing Co. PJS
5.23.112.0	5.23.119.255	Khalij Fars Ettela Resan LTD
5.34.192.0	5.34.192.255	Ayandeh Afzayeh Karaneh Co PJS
5.34.193.0	5.34.193.255	Ayandeh Afzayeh Karaneh Co PJS
5.34.194.0	5.34.195.255	Ayandeh Afzayeh Karaneh Co PJS
5.34.196.0	5.34.199.255	Ayandeh Afzayeh Karaneh Co PJS
5.34.200.0	5.34.200.255	Ayandeh Afzayeh Karaneh Co PJS
5.34.201.0	5.34.201.255	Ayandeh Afzayeh Karaneh Co PJS
5.34.202.0	5.34.203.255	Ayandeh Afzayeh Karaneh Co PJS
5.34.204.0	5.34.205.255	Ayandeh Afzayeh Karaneh Co PJS
5.34.206.0	5.34.206.255	Ayandeh Afzayeh Karaneh Co PJS
5.34.207.0	5.34.207.255	Ayandeh Afzayeh Karaneh Co PJS
5.42.217.0	5.42.217.255	Patron Technology Persia Ltd
5.42.223.0	5.42.223.255	Patron Technology Persia Ltd
5.52.0.0	5.52.0.255	Mobile Communication Company of Iran Plc
5.52.1.0	5.52.1.255	Mobile Communication Company of Iran Plc
5.52.2.0	5.52.3.255	Mobile Communication Company of Iran Plc
5.52.4.0	5.52.7.255	Mobile Communication Company of Iran Plc
5.52.8.0	5.52.15.255	Mobile Communication Company of Iran Plc
5.52.16.0	5.52.31.255	Mobile Communication Company of Iran Plc
5.52.32.0	5.52.35.255	Mobile Communication Company of Iran Plc
5.52.36.0	5.52.36.255	Mobile Communication Company of Iran Plc
5.52.37.0	5.52.37.255	Mobile Communication Company of Iran Plc
5.52.38.0	5.52.39.255	Mobile Communication Company of Iran Plc
5.52.40.0	5.52.47.255	Mobile Communication Company of Iran Plc
5.52.48.0	5.52.63.255	Mobile Communication Company of Iran Plc
5.52.64.0	5.52.79.255	Mobile Communication Company of Iran Plc
5.52.80.0	5.52.95.255	Mobile Communication Company of Iran Plc
5.52.96.0	5.52.111.255	Mobile Communication Company of Iran Plc
5.52.112.0	5.52.127.255	Mobile Communication Company of Iran Plc
5.52.128.0	5.52.128.255	Mobile Communication Company of Iran Plc
5.52.129.0	5.52.129.255	Mobile Communication Company of Iran Plc
5.52.130.0	5.52.131.255	Mobile Communication Company of Iran Plc
5.52.132.0	5.52.135.255	Mobile Communication Company of Iran Plc
5.52.136.0	5.52.143.255	Mobile Communication Company of Iran Plc
5.52.144.0	5.52.159.255	Mobile Communication Company of Iran Plc
5.52.160.0	5.52.175.255	Mobile Communication Company of Iran Plc
5.52.176.0	5.52.191.255	Mobile Communication Company of Iran Plc
5.52.192.0	5.52.207.255	Mobile Communication Company of Iran Plc
5.52.208.0	5.52.223.255	Mobile Communication Company of Iran Plc
5.52.224.0	5.52.239.255	Mobile Communication Company of Iran Plc
5.52.240.0	5.52.255.255	Mobile Communication Company of Iran Plc
5.53.32.0	5.53.63.255	Iran Telecommunication Company Pjs
5.56.128.0	5.56.129.255	Ahvaz Jundishapur University of Medical Sciences
5.56.130.0	5.56.131.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
5.56.132.0	5.56.132.255	Dade Pardazi Mobinhost Co Ltd
5.56.133.0	5.56.133.255	Dade Pardazi Mobinhost Co Ltd
5.56.134.0	5.56.134.255	Aria Web Development LLC
5.56.135.0	5.56.135.255	Dadeh Gostar Sina Shiraz PJSC
5.57.32.0	5.57.32.255	Patron Technology Persia Ltd
5.57.33.0	5.57.33.255	Giti Secure Cloud LLC
5.57.34.0	5.57.34.255	Patron Technology Persia Ltd
5.57.35.0	5.57.35.255	Patron Technology Persia Ltd
5.57.36.0	5.57.36.255	Shabake Tarh Madar Iranian Private Joint Stock Company
5.57.37.0	5.57.37.255	Patron Technology Persia Ltd
5.57.38.0	5.57.38.255	Pars Abr Toseeh Ertebatat Ltd
5.57.39.0	5.57.39.255	Patron Technology Persia Ltd
5.61.24.0	5.61.24.255	Negah Roshan Pars Company (PJS)
5.61.25.0	5.61.25.255	Negah Roshan Pars Company (PJS)
5.61.26.0	5.61.26.255	Negah Roshan Pars Company (PJS)
5.61.28.0	5.61.28.255	Negah Roshan Pars Company (PJS)
5.61.29.0	5.61.29.255	Negah Roshan Pars Company (PJS)
5.61.30.0	5.61.30.255	Negah Roshan Pars Company (PJS)
5.61.31.0	5.61.31.255	Negah Roshan Pars Company (PJS)
5.62.160.0	5.62.191.255	Iran Telecommunication Company Pjs
5.62.192.0	5.62.192.255	Rightel Communication Service Company PJS
5.62.193.0	5.62.193.255	Rightel Communication Service Company PJS
5.62.194.0	5.62.195.255	Rightel Communication Service Company PJS
5.62.196.0	5.62.199.255	Rightel Communication Service Company PJS
5.62.200.0	5.62.207.255	Rightel Communication Service Company PJS
5.62.208.0	5.62.223.255	Rightel Communication Service Company PJS
5.62.224.0	5.62.255.255	Rightel Communication Service Company PJS
5.63.8.0	5.63.8.255	Faraso Samaneh Pasargad Ltd.
5.63.9.0	5.63.9.255	Faraso Samaneh Pasargad Ltd.
5.63.10.0	5.63.10.255	Faraso Samaneh Pasargad Ltd.
5.63.11.0	5.63.11.255	Faraso Samaneh Pasargad Ltd.
5.63.12.0	5.63.12.255	Faraso Samaneh Pasargad Ltd.
5.63.13.0	5.63.13.255	Faraso Samaneh Pasargad Ltd.
5.63.14.0	5.63.14.255	Faraso Samaneh Pasargad Ltd.
5.63.15.0	5.63.15.255	Kermanshah University of Medical Science and Health Service
5.63.23.0	5.63.23.255	Atis Omran Sevin PSJ
5.72.0.0	5.72.31.255	Rightel Communication Service Company PJS
5.72.32.0	5.72.63.255	Rightel Communication Service Company PJS
5.72.64.0	5.72.127.255	Rightel Communication Service Company PJS
5.72.128.0	5.72.255.255	Rightel Communication Service Company PJS
5.73.0.0	5.73.255.255	Rightel Communication Service Company PJS
5.74.0.0	5.74.255.255	Iran Telecommunication Company Pjs
5.75.0.0	5.75.127.255	Iran Telecommunication Company Pjs
5.104.208.0	5.104.215.255	Iran Telecommunication Company Pjs
5.106.0.0	5.106.0.255	Mobile Communication Company of Iran Plc
5.106.1.0	5.106.1.255	Mobile Communication Company of Iran Plc
5.106.2.0	5.106.2.255	Mobile Communication Company of Iran Plc
5.106.3.0	5.106.3.255	Mobile Communication Company of Iran Plc
5.106.4.0	5.106.4.255	Mobile Communication Company of Iran Plc
5.106.5.0	5.106.5.255	Mobile Communication Company of Iran Plc
5.106.6.0	5.106.6.255	Mobile Communication Company of Iran Plc
5.106.7.0	5.106.7.255	Mobile Communication Company of Iran Plc
5.106.8.0	5.106.8.255	Mobile Communication Company of Iran Plc
5.106.9.0	5.106.9.255	Mobile Communication Company of Iran Plc
5.106.10.0	5.106.10.255	Mobile Communication Company of Iran Plc
5.106.11.0	5.106.11.255	Mobile Communication Company of Iran Plc
5.106.12.0	5.106.12.255	Mobile Communication Company of Iran Plc
5.106.13.0	5.106.13.255	Mobile Communication Company of Iran Plc
5.106.14.0	5.106.14.255	Mobile Communication Company of Iran Plc
5.106.15.0	5.106.15.255	Mobile Communication Company of Iran Plc
5.106.16.0	5.106.16.255	Mobile Communication Company of Iran Plc
5.106.17.0	5.106.17.255	Mobile Communication Company of Iran Plc
5.106.18.0	5.106.18.255	Mobile Communication Company of Iran Plc
5.106.19.0	5.106.19.255	Mobile Communication Company of Iran Plc
5.106.20.0	5.106.20.255	Mobile Communication Company of Iran Plc
5.106.21.0	5.106.21.255	Mobile Communication Company of Iran Plc
5.106.22.0	5.106.22.255	Mobile Communication Company of Iran Plc
5.106.23.0	5.106.23.255	Mobile Communication Company of Iran Plc
5.106.24.0	5.106.27.255	Mobile Communication Company of Iran Plc
5.106.28.0	5.106.28.255	Mobile Communication Company of Iran Plc
5.106.29.0	5.106.29.255	Mobile Communication Company of Iran Plc
5.106.30.0	5.106.31.255	Mobile Communication Company of Iran Plc
5.106.32.0	5.106.47.255	Mobile Communication Company of Iran Plc
5.106.48.0	5.106.49.255	Mobile Communication Company of Iran Plc
5.106.50.0	5.106.50.255	Mobile Communication Company of Iran Plc
5.106.51.0	5.106.51.255	Mobile Communication Company of Iran Plc
5.106.52.0	5.106.52.255	Mobile Communication Company of Iran Plc
5.106.53.0	5.106.53.255	Mobile Communication Company of Iran Plc
5.106.54.0	5.106.55.255	Mobile Communication Company of Iran Plc
5.106.56.0	5.106.63.255	Mobile Communication Company of Iran Plc
5.106.64.0	5.106.79.255	Mobile Communication Company of Iran Plc
5.106.80.0	5.106.95.255	Mobile Communication Company of Iran Plc
5.106.96.0	5.106.111.255	Mobile Communication Company of Iran Plc
5.106.112.0	5.106.127.255	Mobile Communication Company of Iran Plc
5.106.128.0	5.106.143.255	Mobile Communication Company of Iran Plc
5.106.144.0	5.106.159.255	Mobile Communication Company of Iran Plc
5.106.160.0	5.106.175.255	Mobile Communication Company of Iran Plc
5.106.176.0	5.106.191.255	Mobile Communication Company of Iran Plc
5.106.192.0	5.106.207.255	Mobile Communication Company of Iran Plc
5.106.208.0	5.106.223.255	Mobile Communication Company of Iran Plc
5.106.224.0	5.106.239.255	Mobile Communication Company of Iran Plc
5.106.240.0	5.106.255.255	Mobile Communication Company of Iran Plc
5.112.0.0	5.112.3.255	Iran Cell Service and Communication Company
5.112.4.0	5.112.5.255	Iran Cell Service and Communication Company
5.112.6.0	5.112.6.255	Iran Cell Service and Communication Company
5.112.7.0	5.112.7.255	Iran Cell Service and Communication Company
5.112.8.0	5.112.15.255	Iran Cell Service and Communication Company
5.112.16.0	5.112.31.255	Iran Cell Service and Communication Company
5.112.32.0	5.112.47.255	Iran Cell Service and Communication Company
5.112.48.0	5.112.63.255	Iran Cell Service and Communication Company
5.112.64.0	5.112.79.255	Iran Cell Service and Communication Company
5.112.80.0	5.112.95.255	Iran Cell Service and Communication Company
5.112.96.0	5.112.111.255	Iran Cell Service and Communication Company
5.112.112.0	5.112.127.255	Iran Cell Service and Communication Company
5.112.128.0	5.112.143.255	Iran Cell Service and Communication Company
5.112.144.0	5.112.159.255	Iran Cell Service and Communication Company
5.112.160.0	5.112.175.255	Iran Cell Service and Communication Company
5.112.176.0	5.112.191.255	Iran Cell Service and Communication Company
5.112.192.0	5.112.207.255	Iran Cell Service and Communication Company
5.112.208.0	5.112.223.255	Iran Cell Service and Communication Company
5.112.224.0	5.112.239.255	Iran Cell Service and Communication Company
5.112.240.0	5.112.255.255	Iran Cell Service and Communication Company
5.113.0.0	5.113.15.255	Iran Cell Service and Communication Company
5.113.16.0	5.113.31.255	Iran Cell Service and Communication Company
5.113.32.0	5.113.47.255	Iran Cell Service and Communication Company
5.113.48.0	5.113.63.255	Iran Cell Service and Communication Company
5.113.64.0	5.113.79.255	Iran Cell Service and Communication Company
5.113.80.0	5.113.81.255	Iran Cell Service and Communication Company
5.113.82.0	5.113.82.255	Iran Cell Service and Communication Company
5.113.83.0	5.113.83.255	Iran Cell Service and Communication Company
5.113.84.0	5.113.87.255	Iran Cell Service and Communication Company
5.113.88.0	5.113.95.255	Iran Cell Service and Communication Company
5.113.96.0	5.113.111.255	Iran Cell Service and Communication Company
5.113.112.0	5.113.127.255	Iran Cell Service and Communication Company
5.113.128.0	5.113.143.255	Iran Cell Service and Communication Company
5.113.144.0	5.113.159.255	Iran Cell Service and Communication Company
5.113.160.0	5.113.175.255	Iran Cell Service and Communication Company
5.113.176.0	5.113.191.255	Iran Cell Service and Communication Company
5.113.192.0	5.113.207.255	Iran Cell Service and Communication Company
5.113.208.0	5.113.223.255	Iran Cell Service and Communication Company
5.113.224.0	5.113.239.255	Iran Cell Service and Communication Company
5.113.240.0	5.113.255.255	Iran Cell Service and Communication Company
5.114.0.0	5.114.15.255	Iran Cell Service and Communication Company
5.114.16.0	5.114.31.255	Iran Cell Service and Communication Company
5.114.32.0	5.114.47.255	Iran Cell Service and Communication Company
5.114.48.0	5.114.63.255	Iran Cell Service and Communication Company
5.114.64.0	5.114.79.255	Iran Cell Service and Communication Company
5.114.80.0	5.114.95.255	Iran Cell Service and Communication Company
5.114.96.0	5.114.103.255	Iran Cell Service and Communication Company
5.114.104.0	5.114.104.255	Iran Cell Service and Communication Company
5.114.105.0	5.114.105.255	Iran Cell Service and Communication Company
5.114.106.0	5.114.107.255	Iran Cell Service and Communication Company
5.114.108.0	5.114.111.255	Iran Cell Service and Communication Company
5.114.112.0	5.114.127.255	Iran Cell Service and Communication Company
5.114.128.0	5.114.143.255	Iran Cell Service and Communication Company
5.114.144.0	5.114.159.255	Iran Cell Service and Communication Company
5.114.160.0	5.114.175.255	Iran Cell Service and Communication Company
5.114.176.0	5.114.191.255	Iran Cell Service and Communication Company
5.114.192.0	5.114.207.255	Iran Cell Service and Communication Company
5.114.208.0	5.114.223.255	Iran Cell Service and Communication Company
5.114.224.0	5.114.239.255	Iran Cell Service and Communication Company
5.114.240.0	5.114.255.255	Iran Cell Service and Communication Company
5.115.0.0	5.115.15.255	Iran Cell Service and Communication Company
5.115.16.0	5.115.31.255	Iran Cell Service and Communication Company
5.115.32.0	5.115.47.255	Iran Cell Service and Communication Company
5.115.48.0	5.115.63.255	Iran Cell Service and Communication Company
5.115.64.0	5.115.79.255	Iran Cell Service and Communication Company
5.115.80.0	5.115.95.255	Iran Cell Service and Communication Company
5.115.96.0	5.115.111.255	Iran Cell Service and Communication Company
5.115.112.0	5.115.127.255	Iran Cell Service and Communication Company
5.115.128.0	5.115.143.255	Iran Cell Service and Communication Company
5.115.144.0	5.115.159.255	Iran Cell Service and Communication Company
5.115.160.0	5.115.175.255	Iran Cell Service and Communication Company
5.115.176.0	5.115.191.255	Iran Cell Service and Communication Company
5.115.192.0	5.115.207.255	Iran Cell Service and Communication Company
5.115.208.0	5.115.223.255	Iran Cell Service and Communication Company
5.115.224.0	5.115.239.255	Iran Cell Service and Communication Company
5.115.240.0	5.115.255.255	Iran Cell Service and Communication Company
5.116.0.0	5.116.15.255	Iran Cell Service and Communication Company
5.116.16.0	5.116.31.255	Iran Cell Service and Communication Company
5.116.32.0	5.116.47.255	Iran Cell Service and Communication Company
5.116.48.0	5.116.63.255	Iran Cell Service and Communication Company
5.116.64.0	5.116.79.255	Iran Cell Service and Communication Company
5.116.80.0	5.116.95.255	Iran Cell Service and Communication Company
5.116.96.0	5.116.111.255	Iran Cell Service and Communication Company
5.116.112.0	5.116.127.255	Iran Cell Service and Communication Company
5.116.128.0	5.116.143.255	Iran Cell Service and Communication Company
5.116.144.0	5.116.159.255	Iran Cell Service and Communication Company
5.116.160.0	5.116.175.255	Iran Cell Service and Communication Company
5.116.176.0	5.116.191.255	Iran Cell Service and Communication Company
5.116.192.0	5.116.207.255	Iran Cell Service and Communication Company
5.116.208.0	5.116.223.255	Iran Cell Service and Communication Company
5.116.224.0	5.116.239.255	Iran Cell Service and Communication Company
5.116.240.0	5.116.255.255	Iran Cell Service and Communication Company
5.117.0.0	5.117.15.255	Iran Cell Service and Communication Company
5.117.16.0	5.117.31.255	Iran Cell Service and Communication Company
5.117.32.0	5.117.47.255	Iran Cell Service and Communication Company
5.117.48.0	5.117.63.255	Iran Cell Service and Communication Company
5.117.64.0	5.117.79.255	Iran Cell Service and Communication Company
5.117.80.0	5.117.95.255	Iran Cell Service and Communication Company
5.117.96.0	5.117.111.255	Iran Cell Service and Communication Company
5.117.112.0	5.117.127.255	Iran Cell Service and Communication Company
5.117.128.0	5.117.143.255	Iran Cell Service and Communication Company
5.117.144.0	5.117.144.255	Iran Cell Service and Communication Company
5.117.145.0	5.117.145.255	Iran Cell Service and Communication Company
5.117.146.0	5.117.147.255	Iran Cell Service and Communication Company
5.117.148.0	5.117.151.255	Iran Cell Service and Communication Company
5.117.152.0	5.117.159.255	Iran Cell Service and Communication Company
5.117.160.0	5.117.175.255	Iran Cell Service and Communication Company
5.117.176.0	5.117.191.255	Iran Cell Service and Communication Company
5.117.192.0	5.117.207.255	Iran Cell Service and Communication Company
5.117.208.0	5.117.223.255	Iran Cell Service and Communication Company
5.117.224.0	5.117.239.255	Iran Cell Service and Communication Company
5.117.240.0	5.117.255.255	Iran Cell Service and Communication Company
5.118.0.0	5.118.255.255	Iran Cell Service and Communication Company
5.119.0.0	5.119.3.255	Iran Cell Service and Communication Company
5.119.4.0	5.119.5.255	Iran Cell Service and Communication Company
5.119.6.0	5.119.6.255	Iran Cell Service and Communication Company
5.119.7.0	5.119.7.255	Iran Cell Service and Communication Company
5.119.8.0	5.119.8.255	Iran Cell Service and Communication Company
5.119.9.0	5.119.9.255	Iran Cell Service and Communication Company
5.119.10.0	5.119.11.255	Iran Cell Service and Communication Company
5.119.12.0	5.119.15.255	Iran Cell Service and Communication Company
5.119.16.0	5.119.31.255	Iran Cell Service and Communication Company
5.119.32.0	5.119.47.255	Iran Cell Service and Communication Company
5.119.48.0	5.119.63.255	Iran Cell Service and Communication Company
5.119.64.0	5.119.67.255	Iran Cell Service and Communication Company
5.119.68.0	5.119.68.255	Iran Cell Service and Communication Company
5.119.69.0	5.119.69.255	Iran Cell Service and Communication Company
5.119.70.0	5.119.71.255	Iran Cell Service and Communication Company
5.119.72.0	5.119.79.255	Iran Cell Service and Communication Company
5.119.80.0	5.119.95.255	Iran Cell Service and Communication Company
5.119.96.0	5.119.111.255	Iran Cell Service and Communication Company
5.119.112.0	5.119.112.255	Iran Cell Service and Communication Company
5.119.113.0	5.119.113.255	Iran Cell Service and Communication Company
5.119.114.0	5.119.115.255	Iran Cell Service and Communication Company
5.119.116.0	5.119.119.255	Iran Cell Service and Communication Company
5.119.120.0	5.119.127.255	Iran Cell Service and Communication Company
5.119.128.0	5.119.143.255	Iran Cell Service and Communication Company
5.119.144.0	5.119.147.255	Iran Cell Service and Communication Company
5.119.148.0	5.119.148.255	Iran Cell Service and Communication Company
5.119.149.0	5.119.149.255	Iran Cell Service and Communication Company
5.119.150.0	5.119.150.255	Iran Cell Service and Communication Company
5.119.151.0	5.119.151.255	Iran Cell Service and Communication Company
5.119.152.0	5.119.155.255	Iran Cell Service and Communication Company
5.119.156.0	5.119.156.255	Iran Cell Service and Communication Company
5.119.157.0	5.119.157.255	Iran Cell Service and Communication Company
5.119.158.0	5.119.159.255	Iran Cell Service and Communication Company
5.119.160.0	5.119.163.255	Iran Cell Service and Communication Company
5.119.164.0	5.119.164.255	Iran Cell Service and Communication Company
5.119.165.0	5.119.165.255	Iran Cell Service and Communication Company
5.119.166.0	5.119.167.255	Iran Cell Service and Communication Company
5.119.168.0	5.119.175.255	Iran Cell Service and Communication Company
5.119.176.0	5.119.191.255	Iran Cell Service and Communication Company
5.119.192.0	5.119.207.255	Iran Cell Service and Communication Company
5.119.208.0	5.119.209.255	Iran Cell Service and Communication Company
5.119.210.0	5.119.210.255	Iran Cell Service and Communication Company
5.119.211.0	5.119.211.255	Iran Cell Service and Communication Company
5.119.212.0	5.119.212.255	Iran Cell Service and Communication Company
5.119.213.0	5.119.213.255	Iran Cell Service and Communication Company
5.119.214.0	5.119.215.255	Iran Cell Service and Communication Company
5.119.216.0	5.119.223.255	Iran Cell Service and Communication Company
5.119.224.0	5.119.239.255	Iran Cell Service and Communication Company
5.119.240.0	5.119.255.255	Iran Cell Service and Communication Company
5.120.0.0	5.120.15.255	Iran Cell Service and Communication Company
5.120.16.0	5.120.31.255	Iran Cell Service and Communication Company
5.120.32.0	5.120.47.255	Iran Cell Service and Communication Company
5.120.48.0	5.120.63.255	Iran Cell Service and Communication Company
5.120.64.0	5.120.79.255	Iran Cell Service and Communication Company
5.120.80.0	5.120.83.255	Iran Cell Service and Communication Company
5.120.84.0	5.120.85.255	Iran Cell Service and Communication Company
5.120.86.0	5.120.86.255	Iran Cell Service and Communication Company
5.120.87.0	5.120.87.255	Iran Cell Service and Communication Company
5.120.88.0	5.120.95.255	Iran Cell Service and Communication Company
5.120.96.0	5.120.111.255	Iran Cell Service and Communication Company
5.120.112.0	5.120.127.255	Iran Cell Service and Communication Company
5.120.128.0	5.120.135.255	Iran Cell Service and Communication Company
5.120.136.0	5.120.137.255	Iran Cell Service and Communication Company
5.120.138.0	5.120.138.255	Iran Cell Service and Communication Company
5.120.139.0	5.120.139.255	Iran Cell Service and Communication Company
5.120.140.0	5.120.143.255	Iran Cell Service and Communication Company
5.120.144.0	5.120.159.255	Iran Cell Service and Communication Company
5.120.160.0	5.120.163.255	Iran Cell Service and Communication Company
5.120.164.0	5.120.164.255	Iran Cell Service and Communication Company
5.120.165.0	5.120.165.255	Iran Cell Service and Communication Company
5.120.166.0	5.120.167.255	Iran Cell Service and Communication Company
5.120.168.0	5.120.175.255	Iran Cell Service and Communication Company
5.120.176.0	5.120.179.255	Iran Cell Service and Communication Company
5.120.180.0	5.120.180.255	Iran Cell Service and Communication Company
5.120.181.0	5.120.181.255	Iran Cell Service and Communication Company
5.120.182.0	5.120.183.255	Iran Cell Service and Communication Company
5.120.184.0	5.120.191.255	Iran Cell Service and Communication Company
5.120.192.0	5.120.207.255	Iran Cell Service and Communication Company
5.120.208.0	5.120.215.255	Iran Cell Service and Communication Company
5.120.216.0	5.120.219.255	Iran Cell Service and Communication Company
5.120.220.0	5.120.221.255	Iran Cell Service and Communication Company
5.120.222.0	5.120.222.255	Iran Cell Service and Communication Company
5.120.223.0	5.120.223.255	Iran Cell Service and Communication Company
5.120.224.0	5.120.239.255	Iran Cell Service and Communication Company
5.120.240.0	5.120.255.255	Iran Cell Service and Communication Company
5.121.0.0	5.121.7.255	Iran Cell Service and Communication Company
5.121.8.0	5.121.11.255	Iran Cell Service and Communication Company
5.121.12.0	5.121.12.255	Iran Cell Service and Communication Company
5.121.13.0	5.121.13.255	Iran Cell Service and Communication Company
5.121.14.0	5.121.15.255	Iran Cell Service and Communication Company
5.121.16.0	5.121.31.255	Iran Cell Service and Communication Company
5.121.32.0	5.121.32.255	Iran Cell Service and Communication Company
5.121.33.0	5.121.33.255	Iran Cell Service and Communication Company
5.121.34.0	5.121.35.255	Iran Cell Service and Communication Company
5.121.36.0	5.121.39.255	Iran Cell Service and Communication Company
5.121.40.0	5.121.47.255	Iran Cell Service and Communication Company
5.121.48.0	5.121.49.255	Iran Cell Service and Communication Company
5.121.50.0	5.121.50.255	Iran Cell Service and Communication Company
5.121.51.0	5.121.51.255	Iran Cell Service and Communication Company
5.121.52.0	5.121.55.255	Iran Cell Service and Communication Company
5.121.56.0	5.121.63.255	Iran Cell Service and Communication Company
5.121.64.0	5.121.79.255	Iran Cell Service and Communication Company
5.121.80.0	5.121.95.255	Iran Cell Service and Communication Company
5.121.96.0	5.121.111.255	Iran Cell Service and Communication Company
5.121.112.0	5.121.127.255	Iran Cell Service and Communication Company
5.121.128.0	5.121.143.255	Iran Cell Service and Communication Company
5.121.144.0	5.121.159.255	Iran Cell Service and Communication Company
5.121.160.0	5.121.175.255	Iran Cell Service and Communication Company
5.121.176.0	5.121.191.255	Iran Cell Service and Communication Company
5.121.192.0	5.121.207.255	Iran Cell Service and Communication Company
5.121.208.0	5.121.223.255	Iran Cell Service and Communication Company
5.121.224.0	5.121.239.255	Iran Cell Service and Communication Company
5.121.240.0	5.121.255.255	Iran Cell Service and Communication Company
5.122.0.0	5.122.15.255	Iran Cell Service and Communication Company
5.122.16.0	5.122.31.255	Iran Cell Service and Communication Company
5.122.32.0	5.122.47.255	Iran Cell Service and Communication Company
5.122.48.0	5.122.63.255	Iran Cell Service and Communication Company
5.122.64.0	5.122.79.255	Iran Cell Service and Communication Company
5.122.80.0	5.122.95.255	Iran Cell Service and Communication Company
5.122.96.0	5.122.111.255	Iran Cell Service and Communication Company
5.122.112.0	5.122.115.255	Iran Cell Service and Communication Company
5.122.116.0	5.122.117.255	Iran Cell Service and Communication Company
5.122.118.0	5.122.118.255	Iran Cell Service and Communication Company
5.122.119.0	5.122.119.255	Iran Cell Service and Communication Company
5.122.120.0	5.122.127.255	Iran Cell Service and Communication Company
5.122.128.0	5.122.143.255	Iran Cell Service and Communication Company
5.122.144.0	5.122.159.255	Iran Cell Service and Communication Company
5.122.160.0	5.122.175.255	Iran Cell Service and Communication Company
5.122.176.0	5.122.179.255	Iran Cell Service and Communication Company
5.122.180.0	5.122.181.255	Iran Cell Service and Communication Company
5.122.182.0	5.122.182.255	Iran Cell Service and Communication Company
5.122.183.0	5.122.183.255	Iran Cell Service and Communication Company
5.122.184.0	5.122.191.255	Iran Cell Service and Communication Company
5.122.192.0	5.122.199.255	Iran Cell Service and Communication Company
5.122.200.0	5.122.201.255	Iran Cell Service and Communication Company
5.122.202.0	5.122.202.255	Iran Cell Service and Communication Company
5.122.203.0	5.122.203.255	Iran Cell Service and Communication Company
5.122.204.0	5.122.207.255	Iran Cell Service and Communication Company
5.122.208.0	5.122.223.255	Iran Cell Service and Communication Company
5.122.224.0	5.122.239.255	Iran Cell Service and Communication Company
5.122.240.0	5.122.255.255	Iran Cell Service and Communication Company
5.123.0.0	5.123.15.255	Iran Cell Service and Communication Company
5.123.16.0	5.123.23.255	Iran Cell Service and Communication Company
5.123.24.0	5.123.27.255	Iran Cell Service and Communication Company
5.123.28.0	5.123.28.255	Iran Cell Service and Communication Company
5.123.29.0	5.123.29.255	Iran Cell Service and Communication Company
5.123.30.0	5.123.31.255	Iran Cell Service and Communication Company
5.123.32.0	5.123.47.255	Iran Cell Service and Communication Company
5.123.48.0	5.123.63.255	Iran Cell Service and Communication Company
5.123.64.0	5.123.79.255	Iran Cell Service and Communication Company
5.123.80.0	5.123.95.255	Iran Cell Service and Communication Company
5.123.96.0	5.123.111.255	Iran Cell Service and Communication Company
5.123.112.0	5.123.127.255	Iran Cell Service and Communication Company
5.123.128.0	5.123.143.255	Iran Cell Service and Communication Company
5.123.144.0	5.123.151.255	Iran Cell Service and Communication Company
5.123.152.0	5.123.152.255	Iran Cell Service and Communication Company
5.123.153.0	5.123.153.255	Iran Cell Service and Communication Company
5.123.154.0	5.123.155.255	Iran Cell Service and Communication Company
5.123.156.0	5.123.159.255	Iran Cell Service and Communication Company
5.123.160.0	5.123.175.255	Iran Cell Service and Communication Company
5.123.176.0	5.123.191.255	Iran Cell Service and Communication Company
5.123.192.0	5.123.207.255	Iran Cell Service and Communication Company
5.123.208.0	5.123.223.255	Iran Cell Service and Communication Company
5.123.224.0	5.123.239.255	Iran Cell Service and Communication Company
5.123.240.0	5.123.255.255	Iran Cell Service and Communication Company
5.124.0.0	5.124.15.255	Iran Cell Service and Communication Company
5.124.16.0	5.124.31.255	Iran Cell Service and Communication Company
5.124.32.0	5.124.32.255	Iran Cell Service and Communication Company
5.124.33.0	5.124.33.255	Iran Cell Service and Communication Company
5.124.34.0	5.124.35.255	Iran Cell Service and Communication Company
5.124.36.0	5.124.39.255	Iran Cell Service and Communication Company
5.124.40.0	5.124.47.255	Iran Cell Service and Communication Company
5.124.48.0	5.124.63.255	Iran Cell Service and Communication Company
5.124.64.0	5.124.79.255	Iran Cell Service and Communication Company
5.124.80.0	5.124.87.255	Iran Cell Service and Communication Company
5.124.88.0	5.124.89.255	Iran Cell Service and Communication Company
5.124.90.0	5.124.90.255	Iran Cell Service and Communication Company
5.124.91.0	5.124.91.255	Iran Cell Service and Communication Company
5.124.92.0	5.124.95.255	Iran Cell Service and Communication Company
5.124.96.0	5.124.111.255	Iran Cell Service and Communication Company
5.124.112.0	5.124.127.255	Iran Cell Service and Communication Company
5.124.128.0	5.124.143.255	Iran Cell Service and Communication Company
5.124.144.0	5.124.159.255	Iran Cell Service and Communication Company
5.124.160.0	5.124.161.255	Iran Cell Service and Communication Company
5.124.162.0	5.124.162.255	Iran Cell Service and Communication Company
5.124.163.0	5.124.163.255	Iran Cell Service and Communication Company
5.124.164.0	5.124.167.255	Iran Cell Service and Communication Company
5.124.168.0	5.124.171.255	Iran Cell Service and Communication Company
5.124.172.0	5.124.173.255	Iran Cell Service and Communication Company
5.124.174.0	5.124.174.255	Iran Cell Service and Communication Company
5.124.175.0	5.124.175.255	Iran Cell Service and Communication Company
5.124.176.0	5.124.191.255	Iran Cell Service and Communication Company
5.124.192.0	5.124.199.255	Iran Cell Service and Communication Company
5.124.200.0	5.124.203.255	Iran Cell Service and Communication Company
5.124.204.0	5.124.204.255	Iran Cell Service and Communication Company
5.124.205.0	5.124.205.255	Iran Cell Service and Communication Company
5.124.206.0	5.124.207.255	Iran Cell Service and Communication Company
5.124.208.0	5.124.215.255	Iran Cell Service and Communication Company
5.124.216.0	5.124.219.255	Iran Cell Service and Communication Company
5.124.220.0	5.124.221.255	Iran Cell Service and Communication Company
5.124.222.0	5.124.222.255	Iran Cell Service and Communication Company
5.124.223.0	5.124.223.255	Iran Cell Service and Communication Company
5.124.224.0	5.124.239.255	Iran Cell Service and Communication Company
5.124.240.0	5.124.255.255	Iran Cell Service and Communication Company
5.125.0.0	5.125.7.255	Iran Cell Service and Communication Company
5.125.8.0	5.125.9.255	Iran Cell Service and Communication Company
5.125.10.0	5.125.10.255	Iran Cell Service and Communication Company
5.125.11.0	5.125.11.255	Iran Cell Service and Communication Company
5.125.12.0	5.125.15.255	Iran Cell Service and Communication Company
5.125.16.0	5.125.31.255	Iran Cell Service and Communication Company
5.125.32.0	5.125.47.255	Iran Cell Service and Communication Company
5.125.48.0	5.125.63.255	Iran Cell Service and Communication Company
5.125.64.0	5.125.79.255	Iran Cell Service and Communication Company
5.125.80.0	5.125.95.255	Iran Cell Service and Communication Company
5.125.96.0	5.125.103.255	Iran Cell Service and Communication Company
5.125.104.0	5.125.104.255	Iran Cell Service and Communication Company
5.125.105.0	5.125.105.255	Iran Cell Service and Communication Company
5.125.106.0	5.125.107.255	Iran Cell Service and Communication Company
5.125.108.0	5.125.111.255	Iran Cell Service and Communication Company
5.125.112.0	5.125.127.255	Iran Cell Service and Communication Company
5.125.128.0	5.125.143.255	Iran Cell Service and Communication Company
5.125.144.0	5.125.159.255	Iran Cell Service and Communication Company
5.125.160.0	5.125.175.255	Iran Cell Service and Communication Company
5.125.176.0	5.125.179.255	Iran Cell Service and Communication Company
5.125.180.0	5.125.180.255	Iran Cell Service and Communication Company
5.125.181.0	5.125.181.255	Iran Cell Service and Communication Company
5.125.182.0	5.125.183.255	Iran Cell Service and Communication Company
5.125.184.0	5.125.191.255	Iran Cell Service and Communication Company
5.125.192.0	5.125.195.255	Iran Cell Service and Communication Company
5.125.196.0	5.125.196.255	Iran Cell Service and Communication Company
5.125.197.0	5.125.197.255	Iran Cell Service and Communication Company
5.125.198.0	5.125.199.255	Iran Cell Service and Communication Company
5.125.200.0	5.125.201.255	Iran Cell Service and Communication Company
5.125.202.0	5.125.202.255	Iran Cell Service and Communication Company
5.125.203.0	5.125.203.255	Iran Cell Service and Communication Company
5.125.204.0	5.125.207.255	Iran Cell Service and Communication Company
5.125.208.0	5.125.211.255	Iran Cell Service and Communication Company
5.125.212.0	5.125.213.255	Iran Cell Service and Communication Company
5.125.214.0	5.125.214.255	Iran Cell Service and Communication Company
5.125.215.0	5.125.215.255	Iran Cell Service and Communication Company
5.125.216.0	5.125.223.255	Iran Cell Service and Communication Company
5.125.224.0	5.125.231.255	Iran Cell Service and Communication Company
5.125.232.0	5.125.233.255	Iran Cell Service and Communication Company
5.125.234.0	5.125.234.255	Iran Cell Service and Communication Company
5.125.235.0	5.125.235.255	Iran Cell Service and Communication Company
5.125.236.0	5.125.239.255	Iran Cell Service and Communication Company
5.125.240.0	5.125.255.255	Iran Cell Service and Communication Company
5.126.0.0	5.126.1.255	Iran Cell Service and Communication Company
5.126.2.0	5.126.2.255	Iran Cell Service and Communication Company
5.126.3.0	5.126.3.255	Iran Cell Service and Communication Company
5.126.4.0	5.126.7.255	Iran Cell Service and Communication Company
5.126.8.0	5.126.15.255	Iran Cell Service and Communication Company
5.126.16.0	5.126.23.255	Iran Cell Service and Communication Company
5.126.24.0	5.126.25.255	Iran Cell Service and Communication Company
5.126.26.0	5.126.26.255	Iran Cell Service and Communication Company
5.126.27.0	5.126.27.255	Iran Cell Service and Communication Company
5.126.28.0	5.126.31.255	Iran Cell Service and Communication Company
5.126.32.0	5.126.47.255	Iran Cell Service and Communication Company
5.126.48.0	5.126.63.255	Iran Cell Service and Communication Company
5.126.64.0	5.126.79.255	Iran Cell Service and Communication Company
5.126.80.0	5.126.80.255	Iran Cell Service and Communication Company
5.126.81.0	5.126.81.255	Iran Cell Service and Communication Company
5.126.82.0	5.126.83.255	Iran Cell Service and Communication Company
5.126.84.0	5.126.87.255	Iran Cell Service and Communication Company
5.126.88.0	5.126.95.255	Iran Cell Service and Communication Company
5.126.96.0	5.126.96.255	Iran Cell Service and Communication Company
5.126.97.0	5.126.97.255	Iran Cell Service and Communication Company
5.126.98.0	5.126.99.255	Iran Cell Service and Communication Company
5.126.100.0	5.126.103.255	Iran Cell Service and Communication Company
5.126.104.0	5.126.111.255	Iran Cell Service and Communication Company
5.126.112.0	5.126.127.255	Iran Cell Service and Communication Company
5.126.128.0	5.126.143.255	Iran Cell Service and Communication Company
5.126.144.0	5.126.151.255	Iran Cell Service and Communication Company
5.126.152.0	5.126.155.255	Iran Cell Service and Communication Company
5.126.156.0	5.126.156.255	Iran Cell Service and Communication Company
5.126.157.0	5.126.157.255	Iran Cell Service and Communication Company
5.126.158.0	5.126.159.255	Iran Cell Service and Communication Company
5.126.160.0	5.126.175.255	Iran Cell Service and Communication Company
5.126.176.0	5.126.191.255	Iran Cell Service and Communication Company
5.126.192.0	5.126.195.255	Iran Cell Service and Communication Company
5.126.196.0	5.126.197.255	Iran Cell Service and Communication Company
5.126.198.0	5.126.198.255	Iran Cell Service and Communication Company
5.126.199.0	5.126.199.255	Iran Cell Service and Communication Company
5.126.200.0	5.126.203.255	Iran Cell Service and Communication Company
5.126.204.0	5.126.204.255	Iran Cell Service and Communication Company
5.126.205.0	5.126.205.255	Iran Cell Service and Communication Company
5.126.206.0	5.126.207.255	Iran Cell Service and Communication Company
5.126.208.0	5.126.209.255	Iran Cell Service and Communication Company
5.126.210.0	5.126.210.255	Iran Cell Service and Communication Company
5.126.211.0	5.126.211.255	Iran Cell Service and Communication Company
5.126.212.0	5.126.212.255	Iran Cell Service and Communication Company
5.126.213.0	5.126.213.255	Iran Cell Service and Communication Company
5.126.214.0	5.126.215.255	Iran Cell Service and Communication Company
5.126.216.0	5.126.219.255	Iran Cell Service and Communication Company
5.126.220.0	5.126.220.255	Iran Cell Service and Communication Company
5.126.221.0	5.126.221.255	Iran Cell Service and Communication Company
5.126.222.0	5.126.223.255	Iran Cell Service and Communication Company
5.126.224.0	5.126.225.255	Iran Cell Service and Communication Company
5.126.226.0	5.126.226.255	Iran Cell Service and Communication Company
5.126.227.0	5.126.227.255	Iran Cell Service and Communication Company
5.126.228.0	5.126.231.255	Iran Cell Service and Communication Company
5.126.232.0	5.126.239.255	Iran Cell Service and Communication Company
5.126.240.0	5.126.243.255	Iran Cell Service and Communication Company
5.126.244.0	5.126.244.255	Iran Cell Service and Communication Company
5.126.245.0	5.126.245.255	Iran Cell Service and Communication Company
5.126.246.0	5.126.247.255	Iran Cell Service and Communication Company
5.126.248.0	5.126.249.255	Iran Cell Service and Communication Company
5.126.250.0	5.126.250.255	Iran Cell Service and Communication Company
5.126.251.0	5.126.251.255	Iran Cell Service and Communication Company
5.126.252.0	5.126.253.255	Iran Cell Service and Communication Company
5.126.254.0	5.126.254.255	Iran Cell Service and Communication Company
5.126.255.0	5.126.255.255	Iran Cell Service and Communication Company
5.127.0.0	5.127.3.255	Iran Cell Service and Communication Company
5.127.4.0	5.127.4.255	Iran Cell Service and Communication Company
5.127.5.0	5.127.5.255	Iran Cell Service and Communication Company
5.127.6.0	5.127.7.255	Iran Cell Service and Communication Company
5.127.8.0	5.127.15.255	Iran Cell Service and Communication Company
5.127.16.0	5.127.17.255	Iran Cell Service and Communication Company
5.127.18.0	5.127.18.255	Iran Cell Service and Communication Company
5.127.19.0	5.127.19.255	Iran Cell Service and Communication Company
5.127.20.0	5.127.23.255	Iran Cell Service and Communication Company
5.127.24.0	5.127.31.255	Iran Cell Service and Communication Company
5.127.32.0	5.127.39.255	Iran Cell Service and Communication Company
5.127.40.0	5.127.41.255	Iran Cell Service and Communication Company
5.127.42.0	5.127.42.255	Iran Cell Service and Communication Company
5.127.43.0	5.127.43.255	Iran Cell Service and Communication Company
5.127.44.0	5.127.45.255	Iran Cell Service and Communication Company
5.127.46.0	5.127.46.255	Iran Cell Service and Communication Company
5.127.47.0	5.127.47.255	Iran Cell Service and Communication Company
5.127.48.0	5.127.55.255	Iran Cell Service and Communication Company
5.127.56.0	5.127.59.255	Iran Cell Service and Communication Company
5.127.60.0	5.127.60.255	Iran Cell Service and Communication Company
5.127.61.0	5.127.61.255	Iran Cell Service and Communication Company
5.127.62.0	5.127.63.255	Iran Cell Service and Communication Company
5.127.64.0	5.127.71.255	Iran Cell Service and Communication Company
5.127.72.0	5.127.75.255	Iran Cell Service and Communication Company
5.127.76.0	5.127.77.255	Iran Cell Service and Communication Company
5.127.78.0	5.127.78.255	Iran Cell Service and Communication Company
5.127.79.0	5.127.79.255	Iran Cell Service and Communication Company
5.127.80.0	5.127.95.255	Iran Cell Service and Communication Company
5.127.96.0	5.127.103.255	Iran Cell Service and Communication Company
5.127.104.0	5.127.107.255	Iran Cell Service and Communication Company
5.127.108.0	5.127.108.255	Iran Cell Service and Communication Company
5.127.109.0	5.127.109.255	Iran Cell Service and Communication Company
5.127.110.0	5.127.110.255	Iran Cell Service and Communication Company
5.127.111.0	5.127.111.255	Iran Cell Service and Communication Company
5.127.112.0	5.127.115.255	Iran Cell Service and Communication Company
5.127.116.0	5.127.117.255	Iran Cell Service and Communication Company
5.127.118.0	5.127.118.255	Iran Cell Service and Communication Company
5.127.119.0	5.127.119.255	Iran Cell Service and Communication Company
5.127.120.0	5.127.127.255	Iran Cell Service and Communication Company
5.127.128.0	5.127.131.255	Iran Cell Service and Communication Company
5.127.132.0	5.127.133.255	Iran Cell Service and Communication Company
5.127.134.0	5.127.134.255	Iran Cell Service and Communication Company
5.127.135.0	5.127.135.255	Iran Cell Service and Communication Company
5.127.136.0	5.127.143.255	Iran Cell Service and Communication Company
5.127.144.0	5.127.159.255	Iran Cell Service and Communication Company
5.127.160.0	5.127.175.255	Iran Cell Service and Communication Company
5.127.176.0	5.127.183.255	Iran Cell Service and Communication Company
5.127.184.0	5.127.187.255	Iran Cell Service and Communication Company
5.127.188.0	5.127.188.255	Iran Cell Service and Communication Company
5.127.189.0	5.127.189.255	Iran Cell Service and Communication Company
5.127.190.0	5.127.191.255	Iran Cell Service and Communication Company
5.127.192.0	5.127.193.255	Iran Cell Service and Communication Company
5.127.194.0	5.127.194.255	Iran Cell Service and Communication Company
5.127.195.0	5.127.195.255	Iran Cell Service and Communication Company
5.127.196.0	5.127.199.255	Iran Cell Service and Communication Company
5.127.200.0	5.127.207.255	Iran Cell Service and Communication Company
5.127.208.0	5.127.211.255	Iran Cell Service and Communication Company
5.127.212.0	5.127.213.255	Iran Cell Service and Communication Company
5.127.214.0	5.127.214.255	Iran Cell Service and Communication Company
5.127.215.0	5.127.215.255	Iran Cell Service and Communication Company
5.127.216.0	5.127.219.255	Iran Cell Service and Communication Company
5.127.220.0	5.127.221.255	Iran Cell Service and Communication Company
5.127.222.0	5.127.222.255	Iran Cell Service and Communication Company
5.127.223.0	5.127.223.255	Iran Cell Service and Communication Company
5.127.224.0	5.127.224.255	Iran Cell Service and Communication Company
5.127.225.0	5.127.225.255	Iran Cell Service and Communication Company
5.127.226.0	5.127.227.255	Iran Cell Service and Communication Company
5.127.228.0	5.127.231.255	Iran Cell Service and Communication Company
5.127.232.0	5.127.239.255	Iran Cell Service and Communication Company
5.127.240.0	5.127.255.255	Iran Cell Service and Communication Company
5.134.128.0	5.134.135.255	Rightel Communication Service Company PJS
5.134.136.0	5.134.139.255	Rightel Communication Service Company PJS
5.134.140.0	5.134.143.255	Rightel Communication Service Company PJS
5.134.144.0	5.134.151.255	Rightel Communication Service Company PJS
5.134.152.0	5.134.155.255	Rightel Communication Service Company PJS
5.134.156.0	5.134.159.255	Rightel Communication Service Company PJS
5.134.160.0	5.134.163.255	Rightel Communication Service Company PJS
5.134.164.0	5.134.167.255	Rightel Communication Service Company PJS
5.134.168.0	5.134.171.255	Rightel Communication Service Company PJS
5.134.172.0	5.134.175.255	Rightel Communication Service Company PJS
5.134.176.0	5.134.179.255	Rightel Communication Service Company PJS
5.134.180.0	5.134.183.255	Rightel Communication Service Company PJS
5.134.184.0	5.134.187.255	Rightel Communication Service Company PJS
5.134.188.0	5.134.189.255	Rightel Communication Service Company PJS
5.134.190.0	5.134.190.255	Rightel Communication Service Company PJS
5.134.191.0	5.134.191.255	Rightel Communication Service Company PJS
5.134.192.0	5.134.195.255	Asiatech Data Transmission Company
5.134.196.0	5.134.199.255	Asiatech Data Transmission Company
5.144.128.0	5.144.128.255	Noavaran Shabakeh Sabz Mehregan (Ltd.)
5.144.129.0	5.144.129.255	Noavaran Shabakeh Sabz Mehregan (Ltd.)
5.144.130.0	5.144.130.255	Noavaran Shabakeh Sabz Mehregan (Ltd.)
5.144.131.0	5.144.131.255	Noavaran Shabakeh Sabz Mehregan (Ltd.)
5.144.132.0	5.144.132.255	Noavaran Shabakeh Sabz Mehregan (Ltd.)
5.144.133.0	5.144.133.255	Noavaran Shabakeh Sabz Mehregan (Ltd.)
5.144.134.0	5.144.134.255	Noavaran Shabakeh Sabz Mehregan (Ltd.)
5.144.135.0	5.144.135.255	Noavaran Shabakeh Sabz Mehregan (Ltd.)
5.145.112.0	5.145.112.255	E-Money Net Developers 24 Company Private Joint Stock
5.145.113.0	5.145.113.255	E-Money Net Developers 24 Company Private Joint Stock
5.145.114.0	5.145.114.255	E-Money Net Developers 24 Company Private Joint Stock
5.145.115.0	5.145.115.255	E-Money Net Developers 24 Company Private Joint Stock
5.145.116.0	5.145.116.255	E-Money Net Developers 24 Company Private Joint Stock
5.145.117.0	5.145.117.255	E-Money Net Developers 24 Company Private Joint Stock
5.145.118.0	5.145.118.255	E-Money Net Developers 24 Company Private Joint Stock
5.145.119.0	5.145.119.255	E-Money Net Developers 24 Company Private Joint Stock
5.159.48.0	5.159.48.255	Padidar Technology PJSC
5.159.49.0	5.159.49.255	Vandad Vira Hooman LLC
5.159.50.0	5.159.50.255	Padidar Technology PJSC
5.159.51.0	5.159.51.255	Padidar Technology PJSC
5.159.52.0	5.159.52.255	Padidar Technology PJSC
5.159.53.0	5.159.53.255	Padidar Technology PJSC
5.159.54.0	5.159.54.255	Ali Monfared
5.159.55.0	5.159.55.255	Padidar Technology PJSC
5.159.192.0	5.159.192.255	Seyed Pouria Mousavizadeh Tehrani
5.160.0.0	5.160.0.255	Respina Networks & Beyond PJSC
5.160.1.0	5.160.1.255	Respina Networks & Beyond PJSC
5.160.2.0	5.160.2.255	Respina Networks & Beyond PJSC
5.160.3.0	5.160.3.255	Respina Networks & Beyond PJSC
5.160.4.0	5.160.4.255	Respina Networks & Beyond PJSC
5.160.5.0	5.160.5.255	Respina Networks & Beyond PJSC
5.160.6.0	5.160.6.255	Respina Networks & Beyond PJSC
5.160.7.0	5.160.7.255	Respina Networks & Beyond PJSC
5.160.8.0	5.160.8.255	Respina Networks & Beyond PJSC
5.160.9.0	5.160.9.255	Respina Networks & Beyond PJSC
5.160.10.0	5.160.10.255	Dade Pardaz Kimia Pouyesh PJS.
5.160.11.0	5.160.11.255	Emdadkhodro Iran
5.160.12.0	5.160.12.255	Respina Networks & Beyond PJSC
5.160.13.0	5.160.13.255	Respina Networks & Beyond PJSC
5.160.14.0	5.160.14.255	Respina Networks & Beyond PJSC
5.160.15.0	5.160.15.255	City Development and Innovation PJSC
5.160.16.0	5.160.17.255	Respina Networks & Beyond PJSC
5.160.18.0	5.160.19.255	Respina Networks & Beyond PJSC
5.160.20.0	5.160.20.255	Respina Networks & Beyond PJSC
5.160.21.0	5.160.21.255	Respina Networks & Beyond PJSC
5.160.22.0	5.160.22.255	Respina Networks & Beyond PJSC
5.160.23.0	5.160.23.255	Respina Networks & Beyond PJSC
5.160.24.0	5.160.24.255	Respina Networks & Beyond PJSC
5.160.25.0	5.160.25.255	Respina Networks & Beyond PJSC
5.160.26.0	5.160.26.255	Respina Networks & Beyond PJSC
5.160.27.0	5.160.27.255	Respina Networks & Beyond PJSC
5.160.28.0	5.160.28.255	Respina Networks & Beyond PJSC
5.160.29.0	5.160.29.255	Respina Networks & Beyond PJSC
5.160.30.0	5.160.30.255	Bimeh Ma PJSC
5.160.31.0	5.160.31.255	Respina Networks & Beyond PJSC
5.160.32.0	5.160.32.255	Respina Networks & Beyond PJSC
5.160.33.0	5.160.33.255	Respina Networks & Beyond PJSC
5.160.34.0	5.160.34.255	Respina Networks & Beyond PJSC
5.160.35.0	5.160.35.255	Respina Networks & Beyond PJSC
5.160.36.0	5.160.36.255	Respina Networks & Beyond PJSC
5.160.37.0	5.160.37.255	Respina Networks & Beyond PJSC
5.160.38.0	5.160.39.255	Respina Networks & Beyond PJSC
5.160.40.0	5.160.40.255	Respina Networks & Beyond PJSC
5.160.41.0	5.160.41.255	Respina Networks & Beyond PJSC
5.160.42.0	5.160.42.255	Respina Networks & Beyond PJSC
5.160.43.0	5.160.43.255	Respina Networks & Beyond PJSC
5.160.44.0	5.160.44.255	Respina Networks & Beyond PJSC
5.160.45.0	5.160.45.255	Respina Networks & Beyond PJSC
5.160.46.0	5.160.46.255	Toesegaran Shabakeh Arseh Novin Ltd
5.160.47.0	5.160.47.255	Rasaneh Avabarid Private Joint Stock Company
5.160.48.0	5.160.48.255	Respina Networks & Beyond PJSC
5.160.49.0	5.160.49.255	Respina Networks & Beyond PJSC
5.160.50.0	5.160.50.255	Respina Networks & Beyond PJSC
5.160.51.0	5.160.51.255	Respina Networks & Beyond PJSC
5.160.52.0	5.160.52.255	Respina Networks & Beyond PJSC
5.160.53.0	5.160.53.255	Respina Networks & Beyond PJSC
5.160.54.0	5.160.54.255	Respina Networks & Beyond PJSC
5.160.55.0	5.160.55.255	Respina Networks & Beyond PJSC
5.160.56.0	5.160.56.255	Respina Networks & Beyond PJSC
5.160.57.0	5.160.57.255	Respina Networks & Beyond PJSC
5.160.58.0	5.160.58.255	Respina Networks & Beyond PJSC
5.160.59.0	5.160.59.255	Respina Networks & Beyond PJSC
5.160.60.0	5.160.60.255	Respina Networks & Beyond PJSC
5.160.61.0	5.160.61.255	Respina Networks & Beyond PJSC
5.160.62.0	5.160.62.255	Respina Networks & Beyond PJSC
5.160.63.0	5.160.63.255	Respina Networks & Beyond PJSC
5.160.64.0	5.160.64.255	Respina Networks & Beyond PJSC
5.160.65.0	5.160.65.255	Respina Networks & Beyond PJSC
5.160.66.0	5.160.66.255	Respina Networks & Beyond PJSC
5.160.67.0	5.160.67.255	Respina Networks & Beyond PJSC
5.160.68.0	5.160.68.255	Respina Networks & Beyond PJSC
5.160.69.0	5.160.69.255	Respina Networks & Beyond PJSC
5.160.70.0	5.160.70.255	Respina Networks & Beyond PJSC
5.160.71.0	5.160.71.255	Respina Networks & Beyond PJSC
5.160.72.0	5.160.72.255	Respina Networks & Beyond PJSC
5.160.73.0	5.160.73.255	Respina Networks & Beyond PJSC
5.160.74.0	5.160.75.255	Respina Networks & Beyond PJSC
5.160.76.0	5.160.76.255	Respina Networks & Beyond PJSC
5.160.77.0	5.160.77.255	Respina Networks & Beyond PJSC
5.160.78.0	5.160.78.255	Respina Networks & Beyond PJSC
5.160.79.0	5.160.79.255	Respina Networks & Beyond PJSC
5.160.80.0	5.160.80.255	Respina Networks & Beyond PJSC
5.160.81.0	5.160.81.255	Respina Networks & Beyond PJSC
5.160.82.0	5.160.82.255	Respina Networks & Beyond PJSC
5.160.83.0	5.160.83.255	Respina Networks & Beyond PJSC
5.160.84.0	5.160.84.255	Respina Networks & Beyond PJSC
5.160.85.0	5.160.85.255	Respina Networks & Beyond PJSC
5.160.86.0	5.160.86.255	Respina Networks & Beyond PJSC
5.160.87.0	5.160.87.255	Respina Networks & Beyond PJSC
5.160.88.0	5.160.88.255	Respina Networks & Beyond PJSC
5.160.89.0	5.160.89.255	Respina Networks & Beyond PJSC
5.160.90.0	5.160.90.255	Respina Networks & Beyond PJSC
5.160.91.0	5.160.91.255	Respina Networks & Beyond PJSC
5.160.92.0	5.160.92.255	Respina Networks & Beyond PJSC
5.160.93.0	5.160.93.255	Respina Networks & Beyond PJSC
5.160.94.0	5.160.94.255	Respina Networks & Beyond PJSC
5.160.95.0	5.160.95.255	Respina Networks & Beyond PJSC
5.160.96.0	5.160.96.255	Respina Networks & Beyond PJSC
5.160.97.0	5.160.97.255	Respina Networks & Beyond PJSC
5.160.98.0	5.160.98.255	Respina Networks & Beyond PJSC
5.160.99.0	5.160.99.255	Respina Networks & Beyond PJSC
5.160.100.0	5.160.100.255	Respina Networks & Beyond PJSC
5.160.101.0	5.160.101.255	Respina Networks & Beyond PJSC
5.160.102.0	5.160.102.255	Respina Networks & Beyond PJSC
5.160.103.0	5.160.103.255	Respina Networks & Beyond PJSC
5.160.104.0	5.160.107.255	Respina Networks & Beyond PJSC
5.160.108.0	5.160.108.255	Respina Networks & Beyond PJSC
5.160.109.0	5.160.109.255	Respina Networks & Beyond PJSC
5.160.110.0	5.160.110.255	Pars Abr Toseeh Ertebatat Ltd
5.160.112.0	5.160.112.255	Respina Networks & Beyond PJSC
5.160.113.0	5.160.113.255	Respina Networks & Beyond PJSC
5.160.114.0	5.160.114.255	Respina Networks & Beyond PJSC
5.160.115.0	5.160.115.255	Respina Networks & Beyond PJSC
5.160.116.0	5.160.116.255	Respina Networks & Beyond PJSC
5.160.117.0	5.160.117.255	Respina Networks & Beyond PJSC
5.160.118.0	5.160.118.255	Respina Networks & Beyond PJSC
5.160.119.0	5.160.119.255	Respina Networks & Beyond PJSC
5.160.120.0	5.160.120.255	Respina Networks & Beyond PJSC
5.160.121.0	5.160.121.255	Respina Networks & Beyond PJSC
5.160.122.0	5.160.122.255	Respina Networks & Beyond PJSC
5.160.123.0	5.160.123.255	Respina Networks & Beyond PJSC
5.160.124.0	5.160.124.255	Respina Networks & Beyond PJSC
5.160.125.0	5.160.125.255	Respina Networks & Beyond PJSC
5.160.126.0	5.160.126.255	Respina Networks & Beyond PJSC
5.160.127.0	5.160.127.255	Respina Networks & Beyond PJSC
5.160.128.0	5.160.128.255	Respina Networks & Beyond PJSC
5.160.129.0	5.160.129.255	Respina Networks & Beyond PJSC
5.160.130.0	5.160.130.255	Respina Networks & Beyond PJSC
5.160.131.0	5.160.131.255	Respina Networks & Beyond PJSC
5.160.132.0	5.160.133.255	Respina Networks & Beyond PJSC
5.160.134.0	5.160.134.255	Respina Networks & Beyond PJSC
5.160.135.0	5.160.135.255	Respina Networks & Beyond PJSC
5.160.136.0	5.160.137.255	Respina Networks & Beyond PJSC
5.160.138.0	5.160.138.255	Respina Networks & Beyond PJSC
5.160.139.0	5.160.139.255	Respina Networks & Beyond PJSC
5.160.140.0	5.160.140.255	Respina Networks & Beyond PJSC
5.160.141.0	5.160.141.255	Respina Networks & Beyond PJSC
5.160.142.0	5.160.142.255	Respina Networks & Beyond PJSC
5.160.143.0	5.160.143.255	Respina Networks & Beyond PJSC
5.160.144.0	5.160.144.255	Pars Abr Toseeh Ertebatat Ltd
5.160.145.0	5.160.145.255	Respina Networks & Beyond PJSC
5.160.146.0	5.160.146.255	Avini Cultural and Art Institute
5.160.147.0	5.160.147.255	Respina Networks & Beyond PJSC
5.160.148.0	5.160.148.255	Respina Networks & Beyond PJSC
5.160.149.0	5.160.149.255	Respina Networks & Beyond PJSC
5.160.150.0	5.160.150.255	Respina Networks & Beyond PJSC
5.160.151.0	5.160.151.255	Respina Networks & Beyond PJSC
5.160.152.0	5.160.152.255	Respina Networks & Beyond PJSC
5.160.153.0	5.160.153.255	Respina Networks & Beyond PJSC
5.160.154.0	5.160.154.255	Resalat Interest-Free Loaning Bank PJS
5.160.155.0	5.160.155.255	Respina Networks & Beyond PJSC
5.160.156.0	5.160.157.255	Respina Networks & Beyond PJSC
5.160.158.0	5.160.158.255	Respina Networks & Beyond PJSC
5.160.159.0	5.160.159.255	Respina Networks & Beyond PJSC
5.160.160.0	5.160.160.255	Pardazeshgar Ray Azma Co. Ltd.
5.160.161.0	5.160.161.255	Pardazeshgar Ray Azma Co. Ltd.
5.160.162.0	5.160.162.255	Respina Networks & Beyond PJSC
5.160.163.0	5.160.163.255	Respina Networks & Beyond PJSC
5.160.164.0	5.160.164.255	Pardazeshgar Ray Azma Co. Ltd.
5.160.165.0	5.160.165.255	Pardazeshgar Ray Azma Co. Ltd.
5.160.166.0	5.160.167.255	Respina Networks & Beyond PJSC
5.160.168.0	5.160.168.255	Respina Networks & Beyond PJSC
5.160.169.0	5.160.169.255	Respina Networks & Beyond PJSC
5.160.170.0	5.160.170.255	Respina Networks & Beyond PJSC
5.160.171.0	5.160.171.255	Respina Networks & Beyond PJSC
5.160.172.0	5.160.172.255	Respina Networks & Beyond PJSC
5.160.173.0	5.160.173.255	Respina Networks & Beyond PJSC
5.160.174.0	5.160.174.255	Respina Networks & Beyond PJSC
5.160.175.0	5.160.175.255	Respina Networks & Beyond PJSC
5.160.176.0	5.160.176.255	Respina Networks & Beyond PJSC
5.160.177.0	5.160.177.255	Respina Networks & Beyond PJSC
5.160.178.0	5.160.178.255	Respina Networks & Beyond PJSC
5.160.179.0	5.160.179.255	Respina Networks & Beyond PJSC
5.160.180.0	5.160.180.255	Pardazeshgar Ray Azma Co. Ltd.
5.160.181.0	5.160.181.255	Pardazeshgar Ray Azma Co. Ltd.
5.160.182.0	5.160.182.255	Pardazeshgar Ray Azma Co. Ltd.
5.160.183.0	5.160.183.255	Pardazeshgar Ray Azma Co. Ltd.
5.160.184.0	5.160.184.255	Respina Networks & Beyond PJSC
5.160.185.0	5.160.185.255	Respina Networks & Beyond PJSC
5.160.186.0	5.160.186.255	Respina Networks & Beyond PJSC
5.160.187.0	5.160.187.255	Respina Networks & Beyond PJSC
5.160.188.0	5.160.188.255	IRIB (Islamic Republic of Iran Broadcasting)
5.160.189.0	5.160.189.255	Respina Networks & Beyond PJSC
5.160.190.0	5.160.190.255	Respina Networks & Beyond PJSC
5.160.191.0	5.160.191.255	Respina Networks & Beyond PJSC
5.160.192.0	5.160.192.255	Saipa Yadak Co. PLC
5.160.193.0	5.160.193.255	Respina Networks & Beyond PJSC
5.160.194.0	5.160.194.255	Industrial Projects Management Company (P.J.S) of Iran
5.160.195.0	5.160.195.255	Road Maintenance & Transportation Organization
5.160.196.0	5.160.196.255	Dade Pardaz Kimia Pouyesh PJS.
5.160.197.0	5.160.197.255	Rasaneh Avabarid Private Joint Stock Company
5.160.198.0	5.160.198.255	Respina Networks & Beyond PJSC
5.160.199.0	5.160.199.255	Shabake Tarh Madar Iranian Private Joint Stock Company
5.160.200.0	5.160.200.255	Respina Networks & Beyond PJSC
5.160.201.0	5.160.201.255	Respina Networks & Beyond PJSC
5.160.202.0	5.160.202.255	Naghsh Awal Keyfiat PJSC
5.160.203.0	5.160.203.255	Naghsh Awal Keyfiat PJSC
5.160.204.0	5.160.207.255	Respina Networks & Beyond PJSC
5.160.208.0	5.160.208.255	Respina Networks & Beyond PJSC
5.160.209.0	5.160.209.255	Respina Networks & Beyond PJSC
5.160.210.0	5.160.210.255	Public JSC Bank Sepah
5.160.211.0	5.160.211.255	Respina Networks & Beyond PJSC
5.160.212.0	5.160.212.255	Respina Networks & Beyond PJSC
5.160.213.0	5.160.213.255	Respina Networks & Beyond PJSC
5.160.214.0	5.160.214.255	Respina Networks & Beyond PJSC
5.160.215.0	5.160.215.255	Respina Networks & Beyond PJSC
5.160.216.0	5.160.216.255	Sarmad Insurance PJSC
5.160.217.0	5.160.217.255	Rahbord Houshmand Shahr PJSC
5.160.218.0	5.160.218.255	Respina Networks & Beyond PJSC
5.160.219.0	5.160.219.255	Respina Networks & Beyond PJSC
5.160.220.0	5.160.220.255	Iran Khodro Co. P.J.S.
5.160.221.0	5.160.221.255	Iran Khodro Co. P.J.S.
5.160.222.0	5.160.222.255	Mahan Air PJSC
5.160.223.0	5.160.223.255	Mahan Air PJSC
5.160.224.0	5.160.224.255	Respina Networks & Beyond PJSC
5.160.225.0	5.160.225.255	Respina Networks & Beyond PJSC
5.160.226.0	5.160.226.255	Respina Networks & Beyond PJSC
5.160.227.0	5.160.227.255	Respina Networks & Beyond PJSC
5.160.228.0	5.160.228.255	Respina Networks & Beyond PJSC
5.160.229.0	5.160.229.255	Respina Networks & Beyond PJSC
5.160.230.0	5.160.230.255	Respina Networks & Beyond PJSC
5.160.231.0	5.160.231.255	Respina Networks & Beyond PJSC
5.160.232.0	5.160.232.255	Respina Networks & Beyond PJSC
5.160.233.0	5.160.233.255	Respina Networks & Beyond PJSC
5.160.234.0	5.160.234.255	Respina Networks & Beyond PJSC
5.160.235.0	5.160.235.255	Respina Networks & Beyond PJSC
5.160.236.0	5.160.236.255	Respina Networks & Beyond PJSC
5.160.237.0	5.160.237.255	Respina Networks & Beyond PJSC
5.160.238.0	5.160.238.255	Respina Networks & Beyond PJSC
5.160.239.0	5.160.239.255	Respina Networks & Beyond PJSC
5.160.240.0	5.160.240.255	Pooya Parto Qeshm Cooperative Company
5.160.241.0	5.160.241.255	Shaparak Electronic Card Payment Network Company PJS
5.160.242.0	5.160.242.255	Bank Refah Kargaran
5.160.243.0	5.160.243.255	Respina Networks & Beyond PJSC
5.160.244.0	5.160.244.255	Middle East Bank (PJSC)
5.160.245.0	5.160.245.255	Middle East Bank (PJSC)
5.160.246.0	5.160.246.255	Respina Networks & Beyond PJSC
5.160.247.0	5.160.247.255	Respina Networks & Beyond PJSC
5.160.248.0	5.160.251.255	Allameh Tabataba'i University
5.160.252.0	5.160.252.255	Respina Networks & Beyond PJSC
5.160.253.0	5.160.253.255	Consumers Cooperative Company of Employees of Esfahan Zobahan Company
5.160.254.0	5.160.254.255	Respina Networks & Beyond PJSC
5.160.255.0	5.160.255.255	Respina Networks & Beyond PJSC
5.182.44.0	5.182.44.255	Green Web Samaneh Novin PJSC
5.182.45.0	5.182.45.255	Green Web Samaneh Novin PJSC
5.182.46.0	5.182.46.255	Green Web Samaneh Novin PJSC
5.190.0.0	5.190.127.255	Iran Telecommunication Company Pjs
5.190.128.0	5.190.191.255	Iran Telecommunication Company Pjs
5.190.192.0	5.190.207.255	Iran Telecommunication Company Pjs
5.190.208.0	5.190.209.255	Iran Telecommunication Company Pjs
5.190.210.0	5.190.210.255	Iran Telecommunication Company Pjs
5.190.211.0	5.190.211.255	Fasa University
5.190.212.0	5.190.215.255	Iran Telecommunication Company Pjs
5.190.216.0	5.190.223.255	Iran Telecommunication Company Pjs
5.190.224.0	5.190.239.255	Iran Telecommunication Company Pjs
5.190.240.0	5.190.255.255	Iran Telecommunication Company Pjs
5.198.160.0	5.198.163.255	Iran Telecommunication Company Pjs
5.198.164.0	5.198.165.255	Iran Telecommunication Company Pjs
5.198.166.0	5.198.167.255	Iran Telecommunication Company Pjs
5.198.168.0	5.198.175.255	Iran Telecommunication Company Pjs
5.198.176.0	5.198.191.255	Iran Telecommunication Company Pjs
5.200.64.0	5.200.64.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
5.200.65.0	5.200.65.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
5.200.66.0	5.200.67.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
5.200.68.0	5.200.71.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
5.200.72.0	5.200.73.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
5.200.74.0	5.200.74.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
5.200.75.0	5.200.75.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
5.200.76.0	5.200.79.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
5.200.80.0	5.200.83.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
5.200.84.0	5.200.84.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
5.200.85.0	5.200.85.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
5.200.86.0	5.200.87.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
5.200.88.0	5.200.95.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
5.200.96.0	5.200.127.255	Rightel Communication Service Company PJS
5.200.128.0	5.200.143.255	Iran Telecommunication Company Pjs
5.200.144.0	5.200.159.255	Iran Telecommunication Company Pjs
5.200.160.0	5.200.175.255	Iran Telecommunication Company Pjs
5.200.176.0	5.200.191.255	Iran Telecommunication Company Pjs
5.200.192.0	5.200.199.255	Iran Telecommunication Company Pjs
5.200.200.0	5.200.200.255	Iran Telecommunication Company Pjs
5.200.201.0	5.200.201.255	Iran Telecommunication Company Pjs
5.200.202.0	5.200.203.255	Iran Telecommunication Company Pjs
5.200.204.0	5.200.207.255	Iran Telecommunication Company Pjs
5.200.208.0	5.200.223.255	Iran Telecommunication Company Pjs
5.200.224.0	5.200.239.255	Iran Telecommunication Company Pjs
5.200.240.0	5.200.247.255	Iran Telecommunication Company Pjs
5.200.248.0	5.200.255.255	Iran Telecommunication Company Pjs
5.201.128.0	5.201.135.255	Mobin Net Communication Company (Private Joint Stock)
5.201.136.0	5.201.137.255	Mobin Net Communication Company (Private Joint Stock)
5.201.138.0	5.201.138.255	Mobin Net Communication Company (Private Joint Stock)
5.201.139.0	5.201.139.255	Mobin Net Communication Company (Private Joint Stock)
5.201.140.0	5.201.141.255	Mobin Net Communication Company (Private Joint Stock)
5.201.142.0	5.201.143.255	Mobin Net Communication Company (Private Joint Stock)
5.201.144.0	5.201.145.255	Mobin Net Communication Company (Private Joint Stock)
5.201.146.0	5.201.147.255	Mobin Net Communication Company (Private Joint Stock)
5.201.148.0	5.201.151.255	Mobin Net Communication Company (Private Joint Stock)
5.201.152.0	5.201.155.255	Mobin Net Communication Company (Private Joint Stock)
5.201.156.0	5.201.159.255	Mobin Net Communication Company (Private Joint Stock)
5.201.160.0	5.201.161.255	Mobin Net Communication Company (Private Joint Stock)
5.201.162.0	5.201.162.255	Mobin Net Communication Company (Private Joint Stock)
5.201.163.0	5.201.163.255	Mobin Net Communication Company (Private Joint Stock)
5.201.164.0	5.201.167.255	Mobin Net Communication Company (Private Joint Stock)
5.201.168.0	5.201.171.255	Mobin Net Communication Company (Private Joint Stock)
5.201.172.0	5.201.173.255	Tehran Municipality ICT Organization
5.201.174.0	5.201.174.255	Mobin Net Communication Company (Private Joint Stock)
5.201.175.0	5.201.175.255	Mobin Net Communication Company (Private Joint Stock)
5.201.176.0	5.201.191.255	Mobin Net Communication Company (Private Joint Stock)
5.201.192.0	5.201.192.255	Mobile Communication Company of Iran Plc
5.201.193.0	5.201.193.255	Mobile Communication Company of Iran Plc
5.201.194.0	5.201.195.255	Mobile Communication Company of Iran Plc
5.201.196.0	5.201.199.255	Mobile Communication Company of Iran Plc
5.201.200.0	5.201.207.255	Mobile Communication Company of Iran Plc
5.201.208.0	5.201.223.255	Mobile Communication Company of Iran Plc
5.201.224.0	5.201.239.255	Mobile Communication Company of Iran Plc
5.201.240.0	5.201.255.255	Mobile Communication Company of Iran Plc
5.202.0.0	5.202.0.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.1.0	5.202.1.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.2.0	5.202.2.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.3.0	5.202.3.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.4.0	5.202.4.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.5.0	5.202.5.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.6.0	5.202.6.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.7.0	5.202.7.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.8.0	5.202.8.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.9.0	5.202.9.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.10.0	5.202.10.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.11.0	5.202.11.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.12.0	5.202.12.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.13.0	5.202.13.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.14.0	5.202.14.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.15.0	5.202.15.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.16.0	5.202.17.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.18.0	5.202.18.255	Didehban Net Company PJS
5.202.19.0	5.202.19.255	Didehban Net Company PJS
5.202.20.0	5.202.20.255	Didehban Net Company PJS
5.202.21.0	5.202.21.255	Didehban Net Company PJS
5.202.22.0	5.202.23.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.24.0	5.202.27.255	Didehban Net Company PJS
5.202.28.0	5.202.28.255	Didehban Net Company PJS
5.202.29.0	5.202.29.255	Didehban Net Company PJS
5.202.30.0	5.202.30.255	Didehban Net Company PJS
5.202.31.0	5.202.31.255	Didehban Net Company PJS
5.202.32.0	5.202.35.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.36.0	5.202.37.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.38.0	5.202.39.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.40.0	5.202.43.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.44.0	5.202.47.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.48.0	5.202.49.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.50.0	5.202.51.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.52.0	5.202.53.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.54.0	5.202.55.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.56.0	5.202.56.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.57.0	5.202.57.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.58.0	5.202.59.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.60.0	5.202.61.255	Didehban Net Company PJS
5.202.62.0	5.202.63.255	Didehban Net Company PJS
5.202.64.0	5.202.67.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.68.0	5.202.68.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.69.0	5.202.69.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.70.0	5.202.71.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.72.0	5.202.75.255	Enteghal Dadeh Mahan Co. PJSC
5.202.76.0	5.202.79.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.80.0	5.202.81.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.82.0	5.202.82.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.83.0	5.202.83.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.84.0	5.202.85.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.86.0	5.202.86.255	Tehran Electric Energy Distribution
5.202.87.0	5.202.87.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.88.0	5.202.95.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.96.0	5.202.99.255	Didehban Net Company PJS
5.202.100.0	5.202.100.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.101.0	5.202.101.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.102.0	5.202.102.255	Enteghal Dadeh Mahan Co. PJSC
5.202.103.0	5.202.103.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.104.0	5.202.105.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.106.0	5.202.106.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.107.0	5.202.107.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.108.0	5.202.109.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.110.0	5.202.110.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.111.0	5.202.111.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.112.0	5.202.112.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.113.0	5.202.113.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.114.0	5.202.115.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.116.0	5.202.119.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.120.0	5.202.123.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.124.0	5.202.125.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.126.0	5.202.127.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.128.0	5.202.128.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.129.0	5.202.129.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.130.0	5.202.131.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.132.0	5.202.135.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.136.0	5.202.143.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.144.0	5.202.159.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.160.0	5.202.163.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.164.0	5.202.165.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.166.0	5.202.167.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.168.0	5.202.169.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.170.0	5.202.171.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.172.0	5.202.172.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.173.0	5.202.173.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.174.0	5.202.175.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.176.0	5.202.177.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.178.0	5.202.179.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.180.0	5.202.180.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.181.0	5.202.181.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.182.0	5.202.183.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.184.0	5.202.185.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.186.0	5.202.186.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.187.0	5.202.187.255	Tosee Fan Afzar Tosan PJSC
5.202.188.0	5.202.189.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.190.0	5.202.191.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.192.0	5.202.199.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.200.0	5.202.203.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.204.0	5.202.205.255	Social Security Organization
5.202.206.0	5.202.207.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.208.0	5.202.223.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.224.0	5.202.231.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.232.0	5.202.235.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.236.0	5.202.239.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.240.0	5.202.243.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.244.0	5.202.244.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.245.0	5.202.245.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.246.0	5.202.247.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.248.0	5.202.251.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.252.0	5.202.252.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.253.0	5.202.253.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.202.254.0	5.202.255.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
5.208.0.0	5.208.15.255	Mobile Communication Company of Iran Plc
5.208.16.0	5.208.31.255	Mobile Communication Company of Iran Plc
5.208.32.0	5.208.47.255	Mobile Communication Company of Iran Plc
5.208.48.0	5.208.63.255	Mobile Communication Company of Iran Plc
5.208.64.0	5.208.79.255	Mobile Communication Company of Iran Plc
5.208.80.0	5.208.95.255	Mobile Communication Company of Iran Plc
5.208.96.0	5.208.111.255	Mobile Communication Company of Iran Plc
5.208.112.0	5.208.127.255	Mobile Communication Company of Iran Plc
5.208.128.0	5.208.143.255	Mobile Communication Company of Iran Plc
5.208.144.0	5.208.159.255	Mobile Communication Company of Iran Plc
5.208.160.0	5.208.175.255	Mobile Communication Company of Iran Plc
5.208.176.0	5.208.191.255	Mobile Communication Company of Iran Plc
5.208.192.0	5.208.207.255	Mobile Communication Company of Iran Plc
5.208.208.0	5.208.223.255	Mobile Communication Company of Iran Plc
5.208.224.0	5.208.239.255	Mobile Communication Company of Iran Plc
5.208.240.0	5.208.255.255	Mobile Communication Company of Iran Plc
5.209.0.0	5.209.15.255	Mobile Communication Company of Iran Plc
5.209.16.0	5.209.31.255	Mobile Communication Company of Iran Plc
5.209.32.0	5.209.47.255	Mobile Communication Company of Iran Plc
5.209.48.0	5.209.63.255	Mobile Communication Company of Iran Plc
5.209.64.0	5.209.79.255	Mobile Communication Company of Iran Plc
5.209.80.0	5.209.95.255	Mobile Communication Company of Iran Plc
5.209.96.0	5.209.111.255	Mobile Communication Company of Iran Plc
5.209.112.0	5.209.127.255	Mobile Communication Company of Iran Plc
5.209.128.0	5.209.143.255	Mobile Communication Company of Iran Plc
5.209.144.0	5.209.159.255	Mobile Communication Company of Iran Plc
5.209.160.0	5.209.175.255	Mobile Communication Company of Iran Plc
5.209.176.0	5.209.191.255	Mobile Communication Company of Iran Plc
5.209.192.0	5.209.207.255	Mobile Communication Company of Iran Plc
5.209.208.0	5.209.223.255	Mobile Communication Company of Iran Plc
5.209.224.0	5.209.239.255	Mobile Communication Company of Iran Plc
5.209.240.0	5.209.255.255	Mobile Communication Company of Iran Plc
5.210.0.0	5.210.0.255	Mobile Communication Company of Iran Plc
5.210.1.0	5.210.1.255	Mobile Communication Company of Iran Plc
5.210.2.0	5.210.3.255	Mobile Communication Company of Iran Plc
5.210.4.0	5.210.7.255	Mobile Communication Company of Iran Plc
5.210.8.0	5.210.15.255	Mobile Communication Company of Iran Plc
5.210.16.0	5.210.31.255	Mobile Communication Company of Iran Plc
5.210.32.0	5.210.47.255	Mobile Communication Company of Iran Plc
5.210.48.0	5.210.63.255	Mobile Communication Company of Iran Plc
5.210.64.0	5.210.79.255	Mobile Communication Company of Iran Plc
5.210.80.0	5.210.95.255	Mobile Communication Company of Iran Plc
5.210.96.0	5.210.111.255	Mobile Communication Company of Iran Plc
5.210.112.0	5.210.127.255	Mobile Communication Company of Iran Plc
5.210.128.0	5.210.143.255	Mobile Communication Company of Iran Plc
5.210.144.0	5.210.159.255	Mobile Communication Company of Iran Plc
5.210.160.0	5.210.175.255	Mobile Communication Company of Iran Plc
5.210.176.0	5.210.191.255	Mobile Communication Company of Iran Plc
5.210.192.0	5.210.207.255	Mobile Communication Company of Iran Plc
5.210.208.0	5.210.223.255	Mobile Communication Company of Iran Plc
5.210.224.0	5.210.239.255	Mobile Communication Company of Iran Plc
5.210.240.0	5.210.255.255	Mobile Communication Company of Iran Plc
5.211.0.0	5.211.15.255	Mobile Communication Company of Iran Plc
5.211.16.0	5.211.31.255	Mobile Communication Company of Iran Plc
5.211.32.0	5.211.47.255	Mobile Communication Company of Iran Plc
5.211.48.0	5.211.63.255	Mobile Communication Company of Iran Plc
5.211.64.0	5.211.79.255	Mobile Communication Company of Iran Plc
5.211.80.0	5.211.95.255	Mobile Communication Company of Iran Plc
5.211.96.0	5.211.111.255	Mobile Communication Company of Iran Plc
5.211.112.0	5.211.127.255	Mobile Communication Company of Iran Plc
5.211.128.0	5.211.128.255	Mobile Communication Company of Iran Plc
5.211.129.0	5.211.129.255	Mobile Communication Company of Iran Plc
5.211.130.0	5.211.131.255	Mobile Communication Company of Iran Plc
5.211.132.0	5.211.135.255	Mobile Communication Company of Iran Plc
5.211.136.0	5.211.143.255	Mobile Communication Company of Iran Plc
5.211.144.0	5.211.159.255	Mobile Communication Company of Iran Plc
5.211.160.0	5.211.175.255	Mobile Communication Company of Iran Plc
5.211.176.0	5.211.191.255	Mobile Communication Company of Iran Plc
5.211.192.0	5.211.207.255	Mobile Communication Company of Iran Plc
5.211.208.0	5.211.223.255	Mobile Communication Company of Iran Plc
5.211.224.0	5.211.239.255	Mobile Communication Company of Iran Plc
5.211.240.0	5.211.255.255	Mobile Communication Company of Iran Plc
5.212.0.0	5.212.0.255	Mobile Communication Company of Iran Plc
5.212.1.0	5.212.1.255	Mobile Communication Company of Iran Plc
5.212.2.0	5.212.3.255	Mobile Communication Company of Iran Plc
5.212.4.0	5.212.7.255	Mobile Communication Company of Iran Plc
5.212.8.0	5.212.15.255	Mobile Communication Company of Iran Plc
5.212.16.0	5.212.31.255	Mobile Communication Company of Iran Plc
5.212.32.0	5.212.47.255	Mobile Communication Company of Iran Plc
5.212.48.0	5.212.63.255	Mobile Communication Company of Iran Plc
5.212.64.0	5.212.79.255	Mobile Communication Company of Iran Plc
5.212.80.0	5.212.95.255	Mobile Communication Company of Iran Plc
5.212.96.0	5.212.111.255	Mobile Communication Company of Iran Plc
5.212.112.0	5.212.127.255	Mobile Communication Company of Iran Plc
5.212.128.0	5.212.143.255	Mobile Communication Company of Iran Plc
5.212.144.0	5.212.159.255	Mobile Communication Company of Iran Plc
5.212.160.0	5.212.175.255	Mobile Communication Company of Iran Plc
5.212.176.0	5.212.191.255	Mobile Communication Company of Iran Plc
5.212.192.0	5.212.207.255	Mobile Communication Company of Iran Plc
5.212.208.0	5.212.223.255	Mobile Communication Company of Iran Plc
5.212.224.0	5.212.239.255	Mobile Communication Company of Iran Plc
5.212.240.0	5.212.255.255	Mobile Communication Company of Iran Plc
5.213.0.0	5.213.0.255	Mobile Communication Company of Iran Plc
5.213.1.0	5.213.1.255	Mobile Communication Company of Iran Plc
5.213.2.0	5.213.3.255	Mobile Communication Company of Iran Plc
5.213.4.0	5.213.7.255	Mobile Communication Company of Iran Plc
5.213.8.0	5.213.15.255	Mobile Communication Company of Iran Plc
5.213.16.0	5.213.31.255	Mobile Communication Company of Iran Plc
5.213.32.0	5.213.47.255	Mobile Communication Company of Iran Plc
5.213.48.0	5.213.63.255	Mobile Communication Company of Iran Plc
5.213.64.0	5.213.79.255	Mobile Communication Company of Iran Plc
5.213.80.0	5.213.95.255	Mobile Communication Company of Iran Plc
5.213.96.0	5.213.111.255	Mobile Communication Company of Iran Plc
5.213.112.0	5.213.127.255	Mobile Communication Company of Iran Plc
5.213.128.0	5.213.128.255	Mobile Communication Company of Iran Plc
5.213.129.0	5.213.129.255	Mobile Communication Company of Iran Plc
5.213.130.0	5.213.131.255	Mobile Communication Company of Iran Plc
5.213.132.0	5.213.135.255	Mobile Communication Company of Iran Plc
5.213.136.0	5.213.143.255	Mobile Communication Company of Iran Plc
5.213.144.0	5.213.159.255	Mobile Communication Company of Iran Plc
5.213.160.0	5.213.175.255	Mobile Communication Company of Iran Plc
5.213.176.0	5.213.191.255	Mobile Communication Company of Iran Plc
5.213.192.0	5.213.192.255	Mobile Communication Company of Iran Plc
5.213.193.0	5.213.193.255	Mobile Communication Company of Iran Plc
5.213.194.0	5.213.195.255	Mobile Communication Company of Iran Plc
5.213.196.0	5.213.199.255	Mobile Communication Company of Iran Plc
5.213.200.0	5.213.207.255	Mobile Communication Company of Iran Plc
5.213.208.0	5.213.223.255	Mobile Communication Company of Iran Plc
5.213.224.0	5.213.239.255	Mobile Communication Company of Iran Plc
5.213.240.0	5.213.255.255	Mobile Communication Company of Iran Plc
5.214.0.0	5.214.0.255	Mobile Communication Company of Iran Plc
5.214.1.0	5.214.1.255	Mobile Communication Company of Iran Plc
5.214.2.0	5.214.3.255	Mobile Communication Company of Iran Plc
5.214.4.0	5.214.7.255	Mobile Communication Company of Iran Plc
5.214.8.0	5.214.15.255	Mobile Communication Company of Iran Plc
5.214.16.0	5.214.31.255	Mobile Communication Company of Iran Plc
5.214.32.0	5.214.47.255	Mobile Communication Company of Iran Plc
5.214.48.0	5.214.63.255	Mobile Communication Company of Iran Plc
5.214.64.0	5.214.79.255	Mobile Communication Company of Iran Plc
5.214.80.0	5.214.95.255	Mobile Communication Company of Iran Plc
5.214.96.0	5.214.111.255	Mobile Communication Company of Iran Plc
5.214.112.0	5.214.127.255	Mobile Communication Company of Iran Plc
5.214.128.0	5.214.128.255	Mobile Communication Company of Iran Plc
5.214.129.0	5.214.129.255	Mobile Communication Company of Iran Plc
5.214.130.0	5.214.131.255	Mobile Communication Company of Iran Plc
5.214.132.0	5.214.135.255	Mobile Communication Company of Iran Plc
5.214.136.0	5.214.143.255	Mobile Communication Company of Iran Plc
5.214.144.0	5.214.159.255	Mobile Communication Company of Iran Plc
5.214.160.0	5.214.175.255	Mobile Communication Company of Iran Plc
5.214.176.0	5.214.191.255	Mobile Communication Company of Iran Plc
5.214.192.0	5.214.192.255	Mobile Communication Company of Iran Plc
5.214.193.0	5.214.193.255	Mobile Communication Company of Iran Plc
5.214.194.0	5.214.195.255	Mobile Communication Company of Iran Plc
5.214.196.0	5.214.199.255	Mobile Communication Company of Iran Plc
5.214.200.0	5.214.207.255	Mobile Communication Company of Iran Plc
5.214.208.0	5.214.223.255	Mobile Communication Company of Iran Plc
5.214.224.0	5.214.239.255	Mobile Communication Company of Iran Plc
5.214.240.0	5.214.255.255	Mobile Communication Company of Iran Plc
5.215.0.0	5.215.0.255	Mobile Communication Company of Iran Plc
5.215.1.0	5.215.1.255	Mobile Communication Company of Iran Plc
5.215.2.0	5.215.3.255	Mobile Communication Company of Iran Plc
5.215.4.0	5.215.7.255	Mobile Communication Company of Iran Plc
5.215.8.0	5.215.15.255	Mobile Communication Company of Iran Plc
5.215.16.0	5.215.31.255	Mobile Communication Company of Iran Plc
5.215.32.0	5.215.47.255	Mobile Communication Company of Iran Plc
5.215.48.0	5.215.63.255	Mobile Communication Company of Iran Plc
5.215.64.0	5.215.79.255	Mobile Communication Company of Iran Plc
5.215.80.0	5.215.95.255	Mobile Communication Company of Iran Plc
5.215.96.0	5.215.111.255	Mobile Communication Company of Iran Plc
5.215.112.0	5.215.127.255	Mobile Communication Company of Iran Plc
5.215.128.0	5.215.128.255	Mobile Communication Company of Iran Plc
5.215.129.0	5.215.129.255	Mobile Communication Company of Iran Plc
5.215.130.0	5.215.131.255	Mobile Communication Company of Iran Plc
5.215.132.0	5.215.135.255	Mobile Communication Company of Iran Plc
5.215.136.0	5.215.143.255	Mobile Communication Company of Iran Plc
5.215.144.0	5.215.159.255	Mobile Communication Company of Iran Plc
5.215.160.0	5.215.175.255	Mobile Communication Company of Iran Plc
5.215.176.0	5.215.191.255	Mobile Communication Company of Iran Plc
5.215.192.0	5.215.207.255	Mobile Communication Company of Iran Plc
5.215.208.0	5.215.223.255	Mobile Communication Company of Iran Plc
5.215.224.0	5.215.239.255	Mobile Communication Company of Iran Plc
5.215.240.0	5.215.255.255	Mobile Communication Company of Iran Plc
5.216.0.0	5.216.0.255	Mobile Communication Company of Iran Plc
5.216.1.0	5.216.1.255	Mobile Communication Company of Iran Plc
5.216.2.0	5.216.3.255	Mobile Communication Company of Iran Plc
5.216.4.0	5.216.7.255	Mobile Communication Company of Iran Plc
5.216.8.0	5.216.15.255	Mobile Communication Company of Iran Plc
5.216.16.0	5.216.31.255	Mobile Communication Company of Iran Plc
5.216.32.0	5.216.47.255	Mobile Communication Company of Iran Plc
5.216.48.0	5.216.63.255	Mobile Communication Company of Iran Plc
5.216.64.0	5.216.79.255	Mobile Communication Company of Iran Plc
5.216.80.0	5.216.95.255	Mobile Communication Company of Iran Plc
5.216.96.0	5.216.111.255	Mobile Communication Company of Iran Plc
5.216.112.0	5.216.127.255	Mobile Communication Company of Iran Plc
5.216.128.0	5.216.128.255	Mobile Communication Company of Iran Plc
5.216.129.0	5.216.129.255	Mobile Communication Company of Iran Plc
5.216.130.0	5.216.131.255	Mobile Communication Company of Iran Plc
5.216.132.0	5.216.135.255	Mobile Communication Company of Iran Plc
5.216.136.0	5.216.143.255	Mobile Communication Company of Iran Plc
5.216.144.0	5.216.159.255	Mobile Communication Company of Iran Plc
5.216.160.0	5.216.175.255	Mobile Communication Company of Iran Plc
5.216.176.0	5.216.191.255	Mobile Communication Company of Iran Plc
5.216.192.0	5.216.207.255	Mobile Communication Company of Iran Plc
5.216.208.0	5.216.223.255	Mobile Communication Company of Iran Plc
5.216.224.0	5.216.239.255	Mobile Communication Company of Iran Plc
5.216.240.0	5.216.255.255	Mobile Communication Company of Iran Plc
5.217.0.0	5.217.0.255	Mobile Communication Company of Iran Plc
5.217.1.0	5.217.1.255	Mobile Communication Company of Iran Plc
5.217.2.0	5.217.3.255	Mobile Communication Company of Iran Plc
5.217.4.0	5.217.7.255	Mobile Communication Company of Iran Plc
5.217.8.0	5.217.15.255	Mobile Communication Company of Iran Plc
5.217.16.0	5.217.31.255	Mobile Communication Company of Iran Plc
5.217.32.0	5.217.47.255	Mobile Communication Company of Iran Plc
5.217.48.0	5.217.63.255	Mobile Communication Company of Iran Plc
5.217.64.0	5.217.79.255	Mobile Communication Company of Iran Plc
5.217.80.0	5.217.95.255	Mobile Communication Company of Iran Plc
5.217.96.0	5.217.111.255	Mobile Communication Company of Iran Plc
5.217.112.0	5.217.127.255	Mobile Communication Company of Iran Plc
5.217.128.0	5.217.128.255	Mobile Communication Company of Iran Plc
5.217.129.0	5.217.129.255	Mobile Communication Company of Iran Plc
5.217.130.0	5.217.131.255	Mobile Communication Company of Iran Plc
5.217.132.0	5.217.135.255	Mobile Communication Company of Iran Plc
5.217.136.0	5.217.143.255	Mobile Communication Company of Iran Plc
5.217.144.0	5.217.159.255	Mobile Communication Company of Iran Plc
5.217.160.0	5.217.175.255	Mobile Communication Company of Iran Plc
5.217.176.0	5.217.191.255	Mobile Communication Company of Iran Plc
5.217.192.0	5.217.207.255	Mobile Communication Company of Iran Plc
5.217.208.0	5.217.223.255	Mobile Communication Company of Iran Plc
5.217.224.0	5.217.239.255	Mobile Communication Company of Iran Plc
5.217.240.0	5.217.255.255	Mobile Communication Company of Iran Plc
5.218.0.0	5.218.15.255	Mobile Communication Company of Iran Plc
5.218.16.0	5.218.31.255	Mobile Communication Company of Iran Plc
5.218.32.0	5.218.47.255	Mobile Communication Company of Iran Plc
5.218.48.0	5.218.63.255	Mobile Communication Company of Iran Plc
5.218.64.0	5.218.79.255	Mobile Communication Company of Iran Plc
5.218.80.0	5.218.95.255	Mobile Communication Company of Iran Plc
5.218.96.0	5.218.111.255	Mobile Communication Company of Iran Plc
5.218.112.0	5.218.127.255	Mobile Communication Company of Iran Plc
5.218.128.0	5.218.143.255	Mobile Communication Company of Iran Plc
5.218.144.0	5.218.159.255	Mobile Communication Company of Iran Plc
5.218.160.0	5.218.175.255	Mobile Communication Company of Iran Plc
5.218.176.0	5.218.191.255	Mobile Communication Company of Iran Plc
5.218.192.0	5.218.207.255	Mobile Communication Company of Iran Plc
5.218.208.0	5.218.223.255	Mobile Communication Company of Iran Plc
5.218.224.0	5.218.239.255	Mobile Communication Company of Iran Plc
5.218.240.0	5.218.255.255	Mobile Communication Company of Iran Plc
5.219.0.0	5.219.63.255	Iran Telecommunication Company Pjs
5.219.64.0	5.219.127.255	Iran Telecommunication Company Pjs
5.219.128.0	5.219.191.255	Iran Telecommunication Company Pjs
5.219.192.0	5.219.255.255	Iran Telecommunication Company Pjs
5.220.0.0	5.221.255.255	GOSTARESH-E-ERTEBATAT-E MABNA COMPANY (Private Joint Stock)
5.232.0.0	5.232.63.255	Iran Telecommunication Company Pjs
5.232.64.0	5.232.95.255	Iran Telecommunication Company Pjs
5.232.96.0	5.232.127.255	Iran Telecommunication Company Pjs
5.232.128.0	5.232.159.255	Iran Telecommunication Company Pjs
5.232.160.0	5.232.175.255	Iran Telecommunication Company Pjs
5.232.176.0	5.232.191.255	Iran Telecommunication Company Pjs
5.232.192.0	5.232.207.255	Iran Telecommunication Company Pjs
5.232.208.0	5.232.215.255	Iran Telecommunication Company Pjs
5.232.216.0	5.232.223.255	Iran Telecommunication Company Pjs
5.232.224.0	5.232.255.255	Iran Telecommunication Company Pjs
5.233.0.0	5.233.31.255	Iran Telecommunication Company Pjs
5.233.32.0	5.233.39.255	Iran Telecommunication Company Pjs
5.233.40.0	5.233.43.255	Iran Telecommunication Company Pjs
5.233.44.0	5.233.45.255	Iran Telecommunication Company Pjs
5.233.46.0	5.233.47.255	Iran Telecommunication Company Pjs
5.233.48.0	5.233.63.255	Iran Telecommunication Company Pjs
5.233.64.0	5.233.95.255	Iran Telecommunication Company Pjs
5.233.96.0	5.233.111.255	Iran Telecommunication Company Pjs
5.233.112.0	5.233.119.255	Iran Telecommunication Company Pjs
5.233.120.0	5.233.127.255	Iran Telecommunication Company Pjs
5.233.128.0	5.233.159.255	Iran Telecommunication Company Pjs
5.233.160.0	5.233.191.255	Iran Telecommunication Company Pjs
5.233.192.0	5.233.223.255	Iran Telecommunication Company Pjs
5.233.224.0	5.233.239.255	Iran Telecommunication Company Pjs
5.233.240.0	5.233.247.255	Iran Telecommunication Company Pjs
5.233.248.0	5.233.249.255	Iran Telecommunication Company Pjs
5.233.250.0	5.233.251.255	Iran Telecommunication Company Pjs
5.233.252.0	5.233.253.255	Iran Telecommunication Company Pjs
5.233.254.0	5.233.255.255	Iran Telecommunication Company Pjs
5.234.0.0	5.234.31.255	Iran Telecommunication Company Pjs
5.234.32.0	5.234.35.255	Iran Telecommunication Company Pjs
5.234.36.0	5.234.39.255	Iran Telecommunication Company Pjs
5.234.40.0	5.234.43.255	Iran Telecommunication Company Pjs
5.234.44.0	5.234.45.255	Iran Telecommunication Company Pjs
5.234.46.0	5.234.47.255	Iran Telecommunication Company Pjs
5.234.48.0	5.234.63.255	Iran Telecommunication Company Pjs
5.234.64.0	5.234.95.255	Iran Telecommunication Company Pjs
5.234.96.0	5.234.127.255	Iran Telecommunication Company Pjs
5.234.128.0	5.234.159.255	Iran Telecommunication Company Pjs
5.234.160.0	5.234.191.255	Iran Telecommunication Company Pjs
5.234.192.0	5.234.207.255	Iran Telecommunication Company Pjs
5.234.208.0	5.234.215.255	Iran Telecommunication Company Pjs
5.234.216.0	5.234.219.255	Iran Telecommunication Company Pjs
5.234.220.0	5.234.223.255	Iran Telecommunication Company Pjs
5.234.224.0	5.234.255.255	Iran Telecommunication Company Pjs
5.235.0.0	5.235.127.255	Iran Telecommunication Company Pjs
5.235.128.0	5.235.143.255	Iran Telecommunication Company Pjs
5.235.144.0	5.235.147.255	Iran Telecommunication Company Pjs
5.235.148.0	5.235.149.255	Iran Telecommunication Company Pjs
5.235.150.0	5.235.151.255	Iran Telecommunication Company Pjs
5.235.152.0	5.235.159.255	Iran Telecommunication Company Pjs
5.235.160.0	5.235.175.255	Iran Telecommunication Company Pjs
5.235.176.0	5.235.183.255	Iran Telecommunication Company Pjs
5.235.184.0	5.235.185.255	Iran Telecommunication Company Pjs
5.235.186.0	5.235.187.255	Iran Telecommunication Company Pjs
5.235.188.0	5.235.191.255	Iran Telecommunication Company Pjs
5.235.192.0	5.235.207.255	Iran Telecommunication Company Pjs
5.235.208.0	5.235.209.255	Iran Telecommunication Company Pjs
5.235.210.0	5.235.210.255	Iran Telecommunication Company Pjs
5.235.211.0	5.235.211.255	Iran Telecommunication Company Pjs
5.235.212.0	5.235.215.255	Iran Telecommunication Company Pjs
5.235.216.0	5.235.223.255	Iran Telecommunication Company Pjs
5.235.224.0	5.235.231.255	Iran Telecommunication Company Pjs
5.235.232.0	5.235.235.255	Iran Telecommunication Company Pjs
5.235.236.0	5.235.239.255	Iran Telecommunication Company Pjs
5.235.240.0	5.235.255.255	Iran Telecommunication Company Pjs
5.236.0.0	5.236.15.255	Iran Telecommunication Company Pjs
5.236.16.0	5.236.23.255	Iran Telecommunication Company Pjs
5.236.24.0	5.236.31.255	Iran Telecommunication Company Pjs
5.236.32.0	5.236.39.255	Iran Telecommunication Company Pjs
5.236.40.0	5.236.47.255	Iran Telecommunication Company Pjs
5.236.48.0	5.236.63.255	Iran Telecommunication Company Pjs
5.236.64.0	5.236.79.255	Iran Telecommunication Company Pjs
5.236.80.0	5.236.87.255	Iran Telecommunication Company Pjs
5.236.88.0	5.236.95.255	Iran Telecommunication Company Pjs
5.236.96.0	5.236.103.255	Iran Telecommunication Company Pjs
5.236.104.0	5.236.105.255	Iran Telecommunication Company Pjs
5.236.106.0	5.236.107.255	Iran Telecommunication Company Pjs
5.236.108.0	5.236.109.255	Iran Telecommunication Company Pjs
5.236.110.0	5.236.111.255	Iran Telecommunication Company Pjs
5.236.112.0	5.236.127.255	Iran Telecommunication Company Pjs
5.236.128.0	5.236.143.255	Iran Telecommunication Company Pjs
5.236.144.0	5.236.151.255	Iran Telecommunication Company Pjs
5.236.152.0	5.236.155.255	Iran Telecommunication Company Pjs
5.236.156.0	5.236.159.255	Iran Telecommunication Company Pjs
5.236.160.0	5.236.191.255	Iran Telecommunication Company Pjs
5.236.192.0	5.236.195.255	Iran Telecommunication Company Pjs
5.236.196.0	5.236.197.255	Iran Telecommunication Company Pjs
5.236.198.0	5.236.199.255	Iran Telecommunication Company Pjs
5.236.200.0	5.236.207.255	Iran Telecommunication Company Pjs
5.236.208.0	5.236.223.255	Iran Telecommunication Company Pjs
5.236.224.0	5.236.255.255	Iran Telecommunication Company Pjs
5.237.0.0	5.237.15.255	Iran Telecommunication Company Pjs
5.237.16.0	5.237.31.255	Iran Telecommunication Company Pjs
5.237.32.0	5.237.47.255	Iran Telecommunication Company Pjs
5.237.48.0	5.237.55.255	Iran Telecommunication Company Pjs
5.237.56.0	5.237.63.255	Iran Telecommunication Company Pjs
5.237.64.0	5.237.71.255	Iran Telecommunication Company Pjs
5.237.72.0	5.237.79.255	Iran Telecommunication Company Pjs
5.237.80.0	5.237.95.255	Iran Telecommunication Company Pjs
5.237.96.0	5.237.127.255	Iran Telecommunication Company Pjs
5.237.128.0	5.237.159.255	Iran Telecommunication Company Pjs
5.237.160.0	5.237.161.255	Iran Telecommunication Company Pjs
5.237.162.0	5.237.163.255	Iran Telecommunication Company Pjs
5.237.164.0	5.237.167.255	Iran Telecommunication Company Pjs
5.237.168.0	5.237.175.255	Iran Telecommunication Company Pjs
5.237.176.0	5.237.191.255	Iran Telecommunication Company Pjs
5.237.192.0	5.237.255.255	Iran Telecommunication Company Pjs
5.238.0.0	5.238.31.255	Iran Telecommunication Company Pjs
5.238.32.0	5.238.47.255	Iran Telecommunication Company Pjs
5.238.48.0	5.238.51.255	Iran Telecommunication Company Pjs
5.238.52.0	5.238.55.255	Iran Telecommunication Company Pjs
5.238.56.0	5.238.63.255	Iran Telecommunication Company Pjs
5.238.64.0	5.238.95.255	Iran Telecommunication Company Pjs
5.238.96.0	5.238.103.255	Iran Telecommunication Company Pjs
5.238.104.0	5.238.107.255	Iran Telecommunication Company Pjs
5.238.108.0	5.238.109.255	Iran Telecommunication Company Pjs
5.238.110.0	5.238.111.255	Iran Telecommunication Company Pjs
5.238.112.0	5.238.127.255	Iran Telecommunication Company Pjs
5.238.128.0	5.238.143.255	Iran Telecommunication Company Pjs
5.238.144.0	5.238.147.255	Iran Telecommunication Company Pjs
5.238.148.0	5.238.151.255	Iran Telecommunication Company Pjs
5.238.152.0	5.238.159.255	Iran Telecommunication Company Pjs
5.238.160.0	5.238.191.255	Iran Telecommunication Company Pjs
5.238.192.0	5.238.255.255	Iran Telecommunication Company Pjs
5.239.0.0	5.239.127.255	Iran Telecommunication Company Pjs
5.239.128.0	5.239.143.255	Iran Telecommunication Company Pjs
5.239.144.0	5.239.159.255	Iran Telecommunication Company Pjs
5.239.160.0	5.239.167.255	Iran Telecommunication Company Pjs
5.239.168.0	5.239.171.255	Iran Telecommunication Company Pjs
5.239.172.0	5.239.175.255	Iran Telecommunication Company Pjs
5.239.176.0	5.239.179.255	Iran Telecommunication Company Pjs
5.239.180.0	5.239.183.255	Iran Telecommunication Company Pjs
5.239.184.0	5.239.191.255	Iran Telecommunication Company Pjs
5.239.192.0	5.239.207.255	Iran Telecommunication Company Pjs
5.239.208.0	5.239.215.255	Iran Telecommunication Company Pjs
5.239.216.0	5.239.223.255	Iran Telecommunication Company Pjs
5.239.224.0	5.239.239.255	Iran Telecommunication Company Pjs
5.239.240.0	5.239.241.255	Iran Telecommunication Company Pjs
5.239.242.0	5.239.243.255	Iran Telecommunication Company Pjs
5.239.244.0	5.239.245.255	Iran Telecommunication Company Pjs
5.239.246.0	5.239.247.255	Iran Telecommunication Company Pjs
5.239.248.0	5.239.255.255	Iran Telecommunication Company Pjs
5.250.0.0	5.250.15.255	Mobile Communication Company of Iran Plc
5.250.16.0	5.250.31.255	Mobile Communication Company of Iran Plc
5.250.32.0	5.250.47.255	Mobile Communication Company of Iran Plc
5.250.48.0	5.250.63.255	Mobile Communication Company of Iran Plc
5.250.64.0	5.250.79.255	Mobile Communication Company of Iran Plc
5.250.80.0	5.250.95.255	Mobile Communication Company of Iran Plc
5.250.96.0	5.250.111.255	Mobile Communication Company of Iran Plc
5.250.112.0	5.250.127.255	Mobile Communication Company of Iran Plc
5.252.216.0	5.252.219.255	Mashhad Municipality Information Technology and Communications Organization
5.253.24.0	5.253.25.255	Noyan Abr Arvan Co. ( Private Joint Stock)
5.253.26.0	5.253.26.255	Noyan Abr Arvan Co. ( Private Joint Stock)
5.253.27.0	5.253.27.255	Noyan Abr Arvan Co. ( Private Joint Stock)
5.253.96.0	5.253.99.255	Moshaverin Parsaye Barin Company Ltd.
5.253.225.0	5.253.225.255	Sindad Network Technology PJSC
14.102.14.0	14.102.14.255	Faraso Samaneh Pasargad Ltd.
31.2.128.0	31.2.143.255	Mobile Communication Company of Iran Plc
31.2.144.0	31.2.159.255	Mobile Communication Company of Iran Plc
31.2.160.0	31.2.175.255	Mobile Communication Company of Iran Plc
31.2.176.0	31.2.191.255	Mobile Communication Company of Iran Plc
31.2.192.0	31.2.192.255	Mobile Communication Company of Iran Plc
31.2.193.0	31.2.193.255	Mobile Communication Company of Iran Plc
31.2.194.0	31.2.195.255	Mobile Communication Company of Iran Plc
31.2.196.0	31.2.199.255	Mobile Communication Company of Iran Plc
31.2.200.0	31.2.207.255	Mobile Communication Company of Iran Plc
31.2.208.0	31.2.223.255	Mobile Communication Company of Iran Plc
31.2.224.0	31.2.239.255	Mobile Communication Company of Iran Plc
31.2.240.0	31.2.255.255	Mobile Communication Company of Iran Plc
31.7.64.0	31.7.65.255	Asiatech Data Transmission Company
31.7.66.0	31.7.66.255	Asiatech Data Transmission Company
31.7.67.0	31.7.67.255	Asiatech Data Transmission Company
31.7.68.0	31.7.69.255	Asiatech Data Transmission Company
31.7.70.0	31.7.70.255	Asiatech Data Transmission Company
31.7.71.0	31.7.71.255	Asiatech Data Transmission Company
31.7.72.0	31.7.72.255	Asiatech Data Transmission Company
31.7.73.0	31.7.73.255	Asiatech Data Transmission Company
31.7.74.0	31.7.75.255	Asiatech Data Transmission Company
31.7.76.0	31.7.77.255	Asiatech Data Transmission Company
31.7.78.0	31.7.78.255	Asiatech Data Transmission Company
31.7.79.0	31.7.79.255	Asiatech Data Transmission Company
31.7.88.0	31.7.88.255	Tarahan Shabake Sharif LTD
31.7.89.0	31.7.89.255	Tarahan Shabake Sharif LTD
31.7.90.0	31.7.90.255	Tarahan Shabake Sharif LTD
31.7.91.0	31.7.91.255	Tarahan Shabake Sharif LTD
31.7.96.0	31.7.103.255	Iran Telecommunication Company Pjs
31.7.104.0	31.7.107.255	Iran Telecommunication Company Pjs
31.7.108.0	31.7.109.255	Iran Telecommunication Company Pjs
31.7.110.0	31.7.111.255	Iran Telecommunication Company Pjs
31.7.112.0	31.7.119.255	Iran Telecommunication Company Pjs
31.7.120.0	31.7.120.255	Iran Telecommunication Company Pjs
31.7.121.0	31.7.121.255	Iran Telecommunication Company Pjs
31.7.122.0	31.7.123.255	Iran Telecommunication Company Pjs
31.7.124.0	31.7.127.255	Iran Telecommunication Company Pjs
31.7.128.0	31.7.128.255	Aryan Satellite Co. (Private Joint Stock)
31.7.129.0	31.7.129.255	Aryan Satellite Co. (Private Joint Stock)
31.7.130.0	31.7.130.255	Aryan Satellite Co. (Private Joint Stock)
31.7.131.0	31.7.131.255	Aryan Satellite Co. (Private Joint Stock)
31.7.132.0	31.7.132.255	Aryan Satellite Co. (Private Joint Stock)
31.7.133.0	31.7.133.255	Aryan Satellite Co. (Private Joint Stock)
31.7.134.0	31.7.134.255	Aryan Satellite Co. (Private Joint Stock)
31.7.135.0	31.7.135.255	Aryan Satellite Co. (Private Joint Stock)
31.7.136.0	31.7.137.255	Aryan Satellite Co. (Private Joint Stock)
31.7.138.0	31.7.138.255	Aryan Satellite Co. (Private Joint Stock)
31.7.139.0	31.7.139.255	Aryan Satellite Co. (Private Joint Stock)
31.7.140.0	31.7.140.255	Aryan Satellite Co. (Private Joint Stock)
31.7.141.0	31.7.141.255	Aryan Satellite Co. (Private Joint Stock)
31.7.142.0	31.7.142.255	Aryan Satellite Co. (Private Joint Stock)
31.7.143.0	31.7.143.255	Aryan Satellite Co. (Private Joint Stock)
31.14.80.0	31.14.95.255	Iran Telecommunication Company Pjs
31.14.112.0	31.14.112.255	Afranet
31.14.113.0	31.14.113.255	Afranet
31.14.114.0	31.14.114.255	Afranet
31.14.115.0	31.14.115.255	Afranet
31.14.116.0	31.14.116.255	Afranet
31.14.117.0	31.14.117.255	Afranet
31.14.118.0	31.14.118.255	Afranet
31.14.119.0	31.14.119.255	Afranet
31.14.120.0	31.14.120.255	Afranet
31.14.121.0	31.14.121.255	Afranet
31.14.122.0	31.14.122.255	Afranet
31.14.123.0	31.14.123.255	Afranet
31.14.124.0	31.14.124.255	Afranet
31.14.125.0	31.14.125.255	Afranet
31.14.126.0	31.14.126.255	Afranet
31.14.127.0	31.14.127.255	Afranet
31.14.144.0	31.14.159.255	Iran Telecommunication Company Pjs
31.24.200.0	31.24.203.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
31.24.204.0	31.24.205.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
31.24.206.0	31.24.206.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
31.24.207.0	31.24.207.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
31.24.232.0	31.24.232.255	Tehran Municipality ICT Organization
31.24.233.0	31.24.233.255	Tehran Municipality ICT Organization
31.24.234.0	31.24.234.255	Tehran Municipality ICT Organization
31.24.235.0	31.24.235.255	Tehran Municipality ICT Organization
31.24.236.0	31.24.239.255	Tehran Municipality ICT Organization
31.25.90.0	31.25.90.255	Saba Abr Mizban LLC
31.25.91.0	31.25.91.255	Saba Abr Mizban LLC
31.25.92.0	31.25.95.255	Cooperative Afra ertebatat-e-sabet-e Rasa Co
31.25.104.0	31.25.111.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
31.25.128.0	31.25.131.255	Asiatech Data Transmission Company
31.25.132.0	31.25.135.255	Asiatech Data Transmission Company
31.40.0.0	31.40.1.255	Mehr Ava Gostar Parsian Information Engineering Co. Ltd
31.40.2.0	31.40.3.255	Mehr Ava Gostar Parsian Information Engineering Co. Ltd
31.40.4.0	31.40.7.255	Mehr Ava Gostar Parsian Information Engineering Co. Ltd
31.41.35.0	31.41.35.255	Padideh Sazan Poshtvar Co. PJS
31.47.32.0	31.47.32.255	Afranet
31.47.33.0	31.47.33.255	Afranet
31.47.34.0	31.47.34.255	Afranet
31.47.35.0	31.47.35.255	Afranet
31.47.36.0	31.47.36.255	Afranet
31.47.37.0	31.47.37.255	Afranet
31.47.38.0	31.47.38.255	Afranet
31.47.39.0	31.47.39.255	Afranet
31.47.40.0	31.47.40.255	Afranet
31.47.41.0	31.47.41.255	Afranet
31.47.42.0	31.47.42.255	Afranet
31.47.43.0	31.47.43.255	Afranet
31.47.44.0	31.47.44.255	Afranet
31.47.45.0	31.47.45.255	Afranet
31.47.46.0	31.47.46.255	Afranet
31.47.47.0	31.47.47.255	Afranet
31.47.48.0	31.47.48.255	Afranet
31.47.49.0	31.47.49.255	Afranet
31.47.50.0	31.47.50.255	Afranet
31.47.51.0	31.47.51.255	Afranet
31.47.52.0	31.47.52.255	Afranet
31.47.53.0	31.47.53.255	Afranet
31.47.54.0	31.47.54.255	Afranet
31.47.55.0	31.47.55.255	Afranet
31.47.56.0	31.47.56.255	Afranet
31.47.57.0	31.47.57.255	Afranet
31.47.58.0	31.47.58.255	Afranet
31.47.59.0	31.47.59.255	Afranet
31.47.60.0	31.47.60.255	Afranet
31.47.61.0	31.47.61.255	Afranet
31.47.62.0	31.47.62.255	Afranet
31.47.63.0	31.47.63.255	Afranet
31.56.89.0	31.56.89.255	Giti Secure Cloud LLC
31.58.237.0	31.58.237.255	Giti Secure Cloud LLC
31.130.176.0	31.130.179.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
31.130.180.0	31.130.180.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
31.130.181.0	31.130.181.255	Atrin Information & Communications Technology Company PJS
31.130.182.0	31.130.182.255	Roshangaran Ertebatat Rayaneh PJSC
31.130.183.0	31.130.183.255	Roshangaran Ertebatat Rayaneh PJSC
31.130.184.0	31.130.191.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
31.170.48.0	31.170.49.255	Farahoosh Dena PLC
31.170.50.0	31.170.51.255	Farahoosh Dena PLC
31.170.52.0	31.170.53.255	Farahoosh Dena PLC
31.170.54.0	31.170.54.255	Farahoosh Dena PLC
31.170.55.0	31.170.55.255	Farahoosh Dena PLC
31.170.56.0	31.170.57.255	Farahoosh Dena PLC
31.170.58.0	31.170.59.255	Farahoosh Dena PLC
31.170.60.0	31.170.63.255	Farahoosh Dena PLC
31.171.216.0	31.171.217.255	Petiak System Co JSC
31.171.218.0	31.171.219.255	Petiak System Co JSC
31.171.220.0	31.171.221.255	Petiak System Co JSC
31.171.222.0	31.171.223.255	Petiak System Co JSC
31.184.128.0	31.184.129.255	GOSTARESH-E-ERTEBATAT-E MABNA COMPANY (Private Joint Stock)
31.184.130.0	31.184.130.255	ICT Research Institute (Iran Telecommunications Research Center Itrc)
31.184.131.0	31.184.131.255	ICT Research Institute (Iran Telecommunications Research Center Itrc)
31.184.132.0	31.184.135.255	GOSTARESH-E-ERTEBATAT-E MABNA COMPANY (Private Joint Stock)
31.184.136.0	31.184.143.255	GOSTARESH-E-ERTEBATAT-E MABNA COMPANY (Private Joint Stock)
31.184.144.0	31.184.144.255	GOSTARESH-E-ERTEBATAT-E MABNA COMPANY (Private Joint Stock)
31.184.145.0	31.184.145.255	Agah Brokerage Company (P.J.S)
31.184.146.0	31.184.147.255	GOSTARESH-E-ERTEBATAT-E MABNA COMPANY (Private Joint Stock)
31.184.148.0	31.184.151.255	GOSTARESH-E-ERTEBATAT-E MABNA COMPANY (Private Joint Stock)
31.184.152.0	31.184.159.255	GOSTARESH-E-ERTEBATAT-E MABNA COMPANY (Private Joint Stock)
31.184.160.0	31.184.163.255	GOSTARESH-E-ERTEBATAT-E MABNA COMPANY (Private Joint Stock)
31.184.164.0	31.184.164.255	Shazand Petrochemical Company (Private Joint Stock)
31.184.165.0	31.184.165.255	GOSTARESH-E-ERTEBATAT-E MABNA COMPANY (Private Joint Stock)
31.184.166.0	31.184.166.255	Mizbani Dadehaye Mabna PJSC
31.184.167.0	31.184.167.255	Mizbani Dadehaye Mabna PJSC
31.184.168.0	31.184.171.255	GOSTARESH-E-ERTEBATAT-E MABNA COMPANY (Private Joint Stock)
31.184.172.0	31.184.172.255	Mizbani Dadehaye Mabna PJSC
31.184.173.0	31.184.173.255	GOSTARESH-E-ERTEBATAT-E MABNA COMPANY (Private Joint Stock)
31.184.174.0	31.184.174.255	GOSTARESH-E-ERTEBATAT-E MABNA COMPANY (Private Joint Stock)
31.184.175.0	31.184.175.255	Mizbani Dadehaye Mabna PJSC
31.184.176.0	31.184.191.255	GOSTARESH-E-ERTEBATAT-E MABNA COMPANY (Private Joint Stock)
31.193.112.0	31.193.119.255	Telecommunication Infrastructure Company
31.193.186.0	31.193.186.255	Sindad Network Technology PJSC
31.214.132.0	31.214.133.255	Parsan Lin Co. PJS
31.214.146.0	31.214.146.255	Parsan Lin Co. PJS
31.214.147.0	31.214.147.255	Parsan Lin Co. PJS
31.214.154.0	31.214.154.255	Parsan Lin Co. PJS
31.214.168.0	31.214.168.255	Parsan Lin Co. PJS
31.214.169.0	31.214.169.255	Parsan Lin Co. PJS
31.214.170.0	31.214.170.255	Parsan Lin Co. PJS
31.214.171.0	31.214.171.255	Parsan Lin Co. PJS
31.214.172.0	31.214.172.255	Parsan Lin Co. PJS
31.214.173.0	31.214.173.255	Parsan Lin Co. PJS
31.214.174.0	31.214.174.255	Parsan Lin Co. PJS
31.214.175.0	31.214.175.255	Parsan Lin Co. PJS
31.214.200.0	31.214.201.255	Parsan Lin Co. PJS
31.214.228.0	31.214.231.255	Parsan Lin Co. PJS
31.214.248.0	31.214.248.255	Parsan Lin Co. PJS
31.214.249.0	31.214.249.255	Parsan Lin Co. PJS
31.214.250.0	31.214.250.255	Parsan Lin Co. PJS
31.214.251.0	31.214.251.255	Parsan Lin Co. PJS
31.214.252.0	31.214.253.255	Parsan Lin Co. PJS
31.214.254.0	31.214.254.255	Parsan Lin Co. PJS
31.214.255.0	31.214.255.255	Parsan Lin Co. PJS
31.216.62.0	31.216.62.255	Asiatech Data Transmission Company
31.217.208.0	31.217.215.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
37.9.248.0	37.9.251.255	Samen Ertebat Asr Co. (P.J.S.)
37.9.252.0	37.9.252.255	Baharan PLC
37.9.253.0	37.9.253.255	Baharan PLC
37.9.254.0	37.9.254.255	Baharan PLC
37.9.255.0	37.9.255.255	Baharan PLC
37.10.64.0	37.10.65.255	Parsan Lin Co. PJS
37.10.66.0	37.10.66.255	Parsan Lin Co. PJS
37.10.67.0	37.10.67.255	Parsan Lin Co. PJS
37.10.109.0	37.10.109.255	Parsan Lin Co. PJS
37.10.117.0	37.10.117.255	Parsan Lin Co. PJS
37.19.80.0	37.19.95.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
37.32.0.0	37.32.0.255	Noyan Abr Arvan Co. ( Private Joint Stock)
37.32.1.0	37.32.1.255	Noyan Abr Arvan Co. ( Private Joint Stock)
37.32.2.0	37.32.2.255	Noyan Abr Arvan Co. ( Private Joint Stock)
37.32.4.0	37.32.7.255	Noyan Abr Arvan Co. ( Private Joint Stock)
37.32.8.0	37.32.11.255	Noyan Abr Arvan Co. ( Private Joint Stock)
37.32.12.0	37.32.15.255	Noyan Abr Arvan Co. ( Private Joint Stock)
37.32.16.0	37.32.16.255	Noyan Abr Arvan Co. ( Private Joint Stock)
37.32.17.0	37.32.17.255	Noyan Abr Arvan Co. ( Private Joint Stock)
37.32.18.0	37.32.18.255	Noyan Abr Arvan Co. ( Private Joint Stock)
37.32.19.0	37.32.19.255	Noyan Abr Arvan Co. ( Private Joint Stock)
37.32.20.0	37.32.23.255	Noyan Abr Arvan Co. ( Private Joint Stock)
37.32.24.0	37.32.27.255	Noyan Abr Arvan Co. ( Private Joint Stock)
37.32.28.0	37.32.31.255	Noyan Abr Arvan Co. ( Private Joint Stock)
37.32.32.0	37.32.35.255	Sefroyek Pardaz Engineering PJSC
37.32.36.0	37.32.39.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
37.32.40.0	37.32.40.255	Bonyan Sepehr Kasir PJSC
37.32.41.0	37.32.41.255	Amir Mohammad Masoudi
37.32.42.0	37.32.42.255	Peyman Ertebatat Pouya Company (Ltd)
37.32.43.0	37.32.43.255	Giti Secure Cloud LLC
37.32.44.0	37.32.45.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
37.32.46.0	37.32.46.255	Enteghal Dade Arya Sarv Co (Ltd.)
37.32.47.0	37.32.47.255	Pars Abr Toseeh Ertebatat Ltd
37.32.112.0	37.32.112.255	Rayaneh Pardazan Baran Co. Ltd.
37.32.113.0	37.32.113.255	Rayaneh Pardazan Baran Co. Ltd.
37.32.114.0	37.32.114.255	Rayaneh Pardazan Baran Co. Ltd.
37.32.115.0	37.32.115.255	Rayaneh Pardazan Baran Co. Ltd.
37.32.116.0	37.32.116.255	Rayaneh Pardazan Baran Co. Ltd.
37.32.117.0	37.32.117.255	Faradast Energy Falat Co. (Private Joint-Stock)
37.32.118.0	37.32.118.255	Rayaneh Pardazan Baran Co. Ltd.
37.32.119.0	37.32.119.255	Rayaneh Pardazan Baran Co. Ltd.
37.32.120.0	37.32.120.255	Rayaneh Pardazan Baran Co. Ltd.
37.32.121.0	37.32.121.255	Rayaneh Pardazan Baran Co. Ltd.
37.32.122.0	37.32.122.255	Rayaneh Pardazan Baran Co. Ltd.
37.32.123.0	37.32.123.255	Rayaneh Pardazan Baran Co. Ltd.
37.32.124.0	37.32.124.255	Rayaneh Pardazan Baran Co. Ltd.
37.32.125.0	37.32.125.255	Rayaneh Pardazan Baran Co. Ltd.
37.32.126.0	37.32.126.255	Rayaneh Pardazan Baran Co. Ltd.
37.32.127.0	37.32.127.255	Rayaneh Pardazan Baran Co. Ltd.
37.44.56.0	37.44.63.255	Rightel Communication Service Company PJS
37.49.145.0	37.49.145.255	Farahoosh Dena PLC
37.49.148.0	37.49.148.255	Farahoosh Dena PLC
37.49.151.0	37.49.151.255	Farahoosh Dena PLC
37.63.128.0	37.63.143.255	Mobile Communication Company of Iran Plc
37.63.144.0	37.63.159.255	Mobile Communication Company of Iran Plc
37.63.160.0	37.63.175.255	Mobile Communication Company of Iran Plc
37.63.176.0	37.63.191.255	Mobile Communication Company of Iran Plc
37.63.192.0	37.63.192.255	Mobile Communication Company of Iran Plc
37.63.193.0	37.63.193.255	Mobile Communication Company of Iran Plc
37.63.194.0	37.63.195.255	Mobile Communication Company of Iran Plc
37.63.196.0	37.63.199.255	Mobile Communication Company of Iran Plc
37.63.200.0	37.63.207.255	Mobile Communication Company of Iran Plc
37.63.208.0	37.63.223.255	Mobile Communication Company of Iran Plc
37.63.224.0	37.63.239.255	Mobile Communication Company of Iran Plc
37.63.240.0	37.63.255.255	Mobile Communication Company of Iran Plc
37.75.240.0	37.75.240.255	Baloot Communication Network Private Limited Company
37.75.241.0	37.75.241.255	Baloot Communication Network Private Limited Company
37.75.242.0	37.75.242.255	Baloot Communication Network Private Limited Company
37.75.243.0	37.75.243.255	Sefroyek Pardaz Engineering PJSC
37.75.244.0	37.75.244.255	Sefroyek Pardaz Engineering PJSC
37.75.245.0	37.75.245.255	Sefroyek Pardaz Engineering PJSC
37.75.246.0	37.75.246.255	Sefroyek Pardaz Engineering PJSC
37.75.247.0	37.75.247.255	Sefroyek Pardaz Engineering PJSC
37.98.0.0	37.98.0.255	Mobile Communication Company of Iran Plc
37.98.1.0	37.98.1.255	Mobile Communication Company of Iran Plc
37.98.2.0	37.98.3.255	Mobile Communication Company of Iran Plc
37.98.4.0	37.98.7.255	Mobile Communication Company of Iran Plc
37.98.8.0	37.98.15.255	Mobile Communication Company of Iran Plc
37.98.16.0	37.98.31.255	Mobile Communication Company of Iran Plc
37.98.32.0	37.98.47.255	Mobile Communication Company of Iran Plc
37.98.48.0	37.98.63.255	Mobile Communication Company of Iran Plc
37.98.64.0	37.98.79.255	Mobile Communication Company of Iran Plc
37.98.80.0	37.98.95.255	Mobile Communication Company of Iran Plc
37.98.96.0	37.98.103.255	Mobile Communication Company of Iran Plc
37.98.104.0	37.98.105.255	Mobile Communication Company of Iran Plc
37.98.106.0	37.98.107.255	Mobile Communication Company of Iran Plc
37.98.108.0	37.98.111.255	Mobile Communication Company of Iran Plc
37.98.112.0	37.98.115.255	Mobin Net Communication Company (Private Joint Stock)
37.98.116.0	37.98.119.255	Mobin Net Communication Company (Private Joint Stock)
37.98.120.0	37.98.127.255	Mobile Communication Company of Iran Plc
37.114.192.0	37.114.199.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
37.114.200.0	37.114.203.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
37.114.204.0	37.114.205.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
37.114.206.0	37.114.207.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
37.114.208.0	37.114.223.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
37.114.224.0	37.114.224.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
37.114.225.0	37.114.225.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
37.114.226.0	37.114.227.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
37.114.228.0	37.114.231.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
37.114.232.0	37.114.239.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
37.114.240.0	37.114.243.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
37.114.244.0	37.114.245.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
37.114.246.0	37.114.247.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
37.114.248.0	37.114.251.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
37.114.252.0	37.114.252.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
37.114.253.0	37.114.253.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
37.114.254.0	37.114.255.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
37.129.0.0	37.129.0.255	Mobile Communication Company of Iran Plc
37.129.1.0	37.129.1.255	Mobile Communication Company of Iran Plc
37.129.2.0	37.129.3.255	Mobile Communication Company of Iran Plc
37.129.4.0	37.129.7.255	Mobile Communication Company of Iran Plc
37.129.8.0	37.129.15.255	Mobile Communication Company of Iran Plc
37.129.16.0	37.129.31.255	Mobile Communication Company of Iran Plc
37.129.32.0	37.129.47.255	Mobile Communication Company of Iran Plc
37.129.48.0	37.129.63.255	Mobile Communication Company of Iran Plc
37.129.64.0	37.129.79.255	Mobile Communication Company of Iran Plc
37.129.80.0	37.129.95.255	Mobile Communication Company of Iran Plc
37.129.96.0	37.129.111.255	Mobile Communication Company of Iran Plc
37.129.112.0	37.129.127.255	Mobile Communication Company of Iran Plc
37.129.128.0	37.129.143.255	Mobile Communication Company of Iran Plc
37.129.144.0	37.129.159.255	Mobile Communication Company of Iran Plc
37.129.160.0	37.129.175.255	Mobile Communication Company of Iran Plc
37.129.176.0	37.129.191.255	Mobile Communication Company of Iran Plc
37.129.192.0	37.129.207.255	Mobile Communication Company of Iran Plc
37.129.208.0	37.129.223.255	Mobile Communication Company of Iran Plc
37.129.224.0	37.129.239.255	Mobile Communication Company of Iran Plc
37.129.240.0	37.129.255.255	Mobile Communication Company of Iran Plc
37.130.200.0	37.130.201.255	GOSTARESH-E-ERTEBATAT-E MABNA COMPANY (Private Joint Stock)
37.130.202.0	37.130.203.255	GOSTARESH-E-ERTEBATAT-E MABNA COMPANY (Private Joint Stock)
37.130.204.0	37.130.204.255	Mizbani Dadehaye Mabna PJSC
37.130.205.0	37.130.205.255	Mizbani Dadehaye Mabna PJSC
37.130.206.0	37.130.206.255	Mizbani Dadehaye Mabna PJSC
37.130.207.0	37.130.207.255	Mizbani Dadehaye Mabna PJSC
37.137.0.0	37.137.31.255	Rightel Communication Service Company PJS
37.137.32.0	37.137.35.255	Rightel Communication Service Company PJS
37.137.36.0	37.137.36.255	Rightel Communication Service Company PJS
37.137.37.0	37.137.37.255	Iranian Presidential Administration
37.137.38.0	37.137.39.255	Rightel Communication Service Company PJS
37.137.40.0	37.137.47.255	Rightel Communication Service Company PJS
37.137.48.0	37.137.63.255	Rightel Communication Service Company PJS
37.137.64.0	37.137.95.255	Rightel Communication Service Company PJS
37.137.96.0	37.137.127.255	Rightel Communication Service Company PJS
37.137.128.0	37.137.255.255	Rightel Communication Service Company PJS
37.143.144.0	37.143.144.255	Rayaneh Gostar Farzanegan Ahwaz Company LTD.
37.143.145.0	37.143.145.255	Rayaneh Gostar Farzanegan Ahwaz Company LTD.
37.143.146.0	37.143.146.255	Rayaneh Gostar Farzanegan Ahwaz Company LTD.
37.143.147.0	37.143.147.255	Rayaneh Gostar Farzanegan Ahwaz Company LTD.
37.143.148.0	37.143.148.255	Rayaneh Gostar Farzanegan Ahwaz Company LTD.
37.143.149.0	37.143.149.255	Rayaneh Gostar Farzanegan Ahwaz Company LTD.
37.143.150.0	37.143.150.255	Rayaneh Gostar Farzanegan Ahwaz Company LTD.
37.143.151.0	37.143.151.255	Rayaneh Gostar Farzanegan Ahwaz Company LTD.
37.148.0.0	37.148.7.255	Aria Shatel PJSC
37.148.8.0	37.148.11.255	Aria Shatel PJSC
37.148.12.0	37.148.15.255	Aria Shatel PJSC
37.148.16.0	37.148.23.255	Aria Shatel PJSC
37.148.24.0	37.148.31.255	Aria Shatel PJSC
37.148.32.0	37.148.39.255	Aria Shatel PJSC
37.148.40.0	37.148.43.255	Aria Shatel PJSC
37.148.44.0	37.148.47.255	Aria Shatel PJSC
37.148.48.0	37.148.49.255	Aria Shatel PJSC
37.148.50.0	37.148.51.255	Aria Shatel PJSC
37.148.52.0	37.148.55.255	Aria Shatel PJSC
37.148.56.0	37.148.63.255	Aria Shatel PJSC
37.148.64.0	37.148.67.255	Aria Shatel PJSC
37.148.68.0	37.148.69.255	Aria Shatel PJSC
37.148.70.0	37.148.71.255	Aria Shatel PJSC
37.148.72.0	37.148.79.255	Aria Shatel PJSC
37.148.80.0	37.148.83.255	Aria Shatel PJSC
37.148.84.0	37.148.85.255	Aria Shatel PJSC
37.148.86.0	37.148.87.255	Aria Shatel PJSC
37.148.88.0	37.148.88.255	Aria Shatel PJSC
37.148.89.0	37.148.89.255	Aria Shatel PJSC
37.148.90.0	37.148.90.255	Aria Shatel PJSC
37.148.91.0	37.148.91.255	Aria Shatel PJSC
37.148.92.0	37.148.92.255	Aria Shatel PJSC
37.148.93.0	37.148.93.255	Aria Shatel PJSC
37.148.94.0	37.148.95.255	Aria Shatel PJSC
37.148.96.0	37.148.111.255	Aria Shatel PJSC
37.148.112.0	37.148.127.255	Aria Shatel PJSC
37.148.248.0	37.148.251.255	Sina Financial & Investment Holding Co (PJS)
37.152.160.0	37.152.160.255	Rahanet Zanjan Co. (Private Joint-Stock)
37.152.161.0	37.152.161.255	Rahanet Zanjan Co. (Private Joint-Stock)
37.152.162.0	37.152.162.255	Rahanet Zanjan Co. (Private Joint-Stock)
37.152.163.0	37.152.163.255	Rahanet Zanjan Co. (Private Joint-Stock)
37.152.164.0	37.152.164.255	Rahanet Zanjan Co. (Private Joint-Stock)
37.152.165.0	37.152.165.255	Pardazeshgaran Etelat Zaban Layeha Co Ltd
37.152.166.0	37.152.166.255	Pardazeshgaran Etelat Zaban Layeha Co Ltd
37.152.167.0	37.152.167.255	Rahanet Zanjan Co. (Private Joint-Stock)
37.152.168.0	37.152.168.255	Rahanet Zanjan Co. (Private Joint-Stock)
37.152.169.0	37.152.169.255	Rahanet Zanjan Co. (Private Joint-Stock)
37.152.170.0	37.152.170.255	Rahanet Zanjan Co. (Private Joint-Stock)
37.152.171.0	37.152.171.255	Rahanet Zanjan Co. (Private Joint-Stock)
37.152.172.0	37.152.172.255	Rahanet Zanjan Co. (Private Joint-Stock)
37.152.176.0	37.152.183.255	Noyan Abr Arvan Co. ( Private Joint Stock)
37.152.184.0	37.152.184.255	Noyan Abr Arvan Co. ( Private Joint Stock)
37.152.185.0	37.152.185.255	Noyan Abr Arvan Co. ( Private Joint Stock)
37.152.186.0	37.152.187.255	Noyan Abr Arvan Co. ( Private Joint Stock)
37.152.188.0	37.152.191.255	Noyan Abr Arvan Co. ( Private Joint Stock)
37.153.128.0	37.153.131.255	Iran Telecommunication Company Pjs
37.153.176.0	37.153.191.255	Rightel Communication Service Company PJS
37.156.0.0	37.156.3.255	Iran Telecommunication Company Pjs
37.156.8.0	37.156.11.255	Asiatech Data Transmission Company
37.156.12.0	37.156.15.255	Asiatech Data Transmission Company
37.156.16.0	37.156.23.255	Mobin Net Communication Company (Private Joint Stock)
37.156.24.0	37.156.27.255	Mobin Net Communication Company (Private Joint Stock)
37.156.28.0	37.156.29.255	Mobin Net Communication Company (Private Joint Stock)
37.156.30.0	37.156.31.255	Mobin Net Communication Company (Private Joint Stock)
37.156.48.0	37.156.63.255	Rightel Communication Service Company PJS
37.156.100.0	37.156.103.255	Iran Telecommunication Company Pjs
37.156.112.0	37.156.127.255	Iran Telecommunication Company Pjs
37.156.128.0	37.156.143.255	Iran Telecommunication Company Pjs
37.156.144.0	37.156.144.255	Asiatech Data Transmission Company
37.156.145.0	37.156.145.255	Asiatech Data Transmission Company
37.156.146.0	37.156.146.255	Asiatech Data Transmission Company
37.156.147.0	37.156.147.255	Asiatech Data Transmission Company
37.156.152.0	37.156.159.255	Rightel Communication Service Company PJS
37.156.160.0	37.156.167.255	Iran Telecommunication Company Pjs
37.156.176.0	37.156.179.255	Iran Telecommunication Company Pjs
37.156.212.0	37.156.215.255	Iran Telecommunication Company Pjs
37.156.232.0	37.156.239.255	Iran Telecommunication Company Pjs
37.156.240.0	37.156.243.255	Iran Telecommunication Company Pjs
37.156.248.0	37.156.251.255	Iran Telecommunication Company Pjs
37.191.64.0	37.191.67.255	Ariana Gostar Spadana (PJSC)
37.191.68.0	37.191.71.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
37.191.72.0	37.191.73.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
37.191.74.0	37.191.74.255	Ariana Gostar Spadana (PJSC)
37.191.75.0	37.191.75.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
37.191.76.0	37.191.79.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
37.191.80.0	37.191.83.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
37.191.84.0	37.191.85.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
37.191.86.0	37.191.87.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
37.191.88.0	37.191.91.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
37.191.92.0	37.191.95.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
37.202.128.0	37.202.135.255	Aria Shatel PJSC
37.202.136.0	37.202.143.255	Aria Shatel PJSC
37.202.144.0	37.202.147.255	Aria Shatel PJSC
37.202.148.0	37.202.149.255	Aria Shatel PJSC
37.202.150.0	37.202.151.255	Aria Shatel PJSC
37.202.152.0	37.202.159.255	Aria Shatel PJSC
37.202.160.0	37.202.167.255	Aria Shatel PJSC
37.202.168.0	37.202.175.255	Aria Shatel PJSC
37.202.176.0	37.202.179.255	Aria Shatel PJSC
37.202.180.0	37.202.183.255	Aria Shatel PJSC
37.202.184.0	37.202.187.255	Aria Shatel PJSC
37.202.188.0	37.202.191.255	Aria Shatel PJSC
37.202.224.0	37.202.224.255	Fanavaran Mehr Vatan Tehran Server Group LLC
37.202.225.0	37.202.225.255	Fanavaran Mehr Vatan Tehran Server Group LLC
37.202.226.0	37.202.226.255	Fanavaran Mehr Vatan Tehran Server Group LLC
37.202.227.0	37.202.227.255	Fanavaran Mehr Vatan Tehran Server Group LLC
37.202.228.0	37.202.228.255	Fanavaran Mehr Vatan Tehran Server Group LLC
37.202.229.0	37.202.229.255	Fanavaran Mehr Vatan Tehran Server Group LLC
37.202.230.0	37.202.230.255	Fanavaran Mehr Vatan Tehran Server Group LLC
37.202.231.0	37.202.231.255	Fanavaran Mehr Vatan Tehran Server Group LLC
37.202.232.0	37.202.232.255	Fanavaran Mehr Vatan Tehran Server Group LLC
37.202.233.0	37.202.233.255	Fanavaran Mehr Vatan Tehran Server Group LLC
37.202.234.0	37.202.234.255	Fanavaran Mehr Vatan Tehran Server Group LLC
37.202.235.0	37.202.235.255	Fanavaran Mehr Vatan Tehran Server Group LLC
37.202.236.0	37.202.236.255	Fanavaran Mehr Vatan Tehran Server Group LLC
37.202.237.0	37.202.237.255	Fanavaran Mehr Vatan Tehran Server Group LLC
37.202.238.0	37.202.238.255	Fanavaran Mehr Vatan Tehran Server Group LLC
37.202.239.0	37.202.239.255	Fanavaran Mehr Vatan Tehran Server Group LLC
37.202.240.0	37.202.240.255	Fanavaran Mehr Vatan Tehran Server Group LLC
37.202.241.0	37.202.241.255	Fanavaran Mehr Vatan Tehran Server Group LLC
37.202.242.0	37.202.242.255	Fanavaran Mehr Vatan Tehran Server Group LLC
37.202.243.0	37.202.243.255	QOM Municipality
37.202.244.0	37.202.244.255	Fanavaran Mehr Vatan Tehran Server Group LLC
37.202.245.0	37.202.245.255	Aria Shatel PJSC
37.202.246.0	37.202.246.255	Fanavaran Mehr Vatan Tehran Server Group LLC
37.202.247.0	37.202.247.255	Fanavaran Mehr Vatan Tehran Server Group LLC
37.202.248.0	37.202.249.255	Aria Shatel PJSC
37.202.250.0	37.202.250.255	Aria Shatel PJSC
37.202.251.0	37.202.251.255	Fanavaran Mehr Vatan Tehran Server Group LLC
37.202.252.0	37.202.255.255	Aria Shatel PJSC
37.221.0.0	37.221.15.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
37.221.16.0	37.221.31.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
37.221.32.0	37.221.63.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
37.228.131.0	37.228.131.255	Parsan Lin Co. PJS
37.228.133.0	37.228.133.255	Parsan Lin Co. PJS
37.228.135.0	37.228.135.255	Parsan Lin Co. PJS
37.228.136.0	37.228.136.255	Parsan Lin Co. PJS
37.228.137.0	37.228.137.255	Parsan Lin Co. PJS
37.228.138.0	37.228.138.255	Parsan Lin Co. PJS
37.228.139.0	37.228.139.255	Parsan Lin Co. PJS
37.235.16.0	37.235.31.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
37.254.0.0	37.254.63.255	Iran Telecommunication Company Pjs
37.254.64.0	37.254.95.255	Iran Telecommunication Company Pjs
37.254.96.0	37.254.127.255	Iran Telecommunication Company Pjs
37.254.128.0	37.254.191.255	Iran Telecommunication Company Pjs
37.254.192.0	37.254.255.255	Iran Telecommunication Company Pjs
37.255.0.0	37.255.63.255	Iran Telecommunication Company Pjs
37.255.64.0	37.255.127.255	Iran Telecommunication Company Pjs
37.255.128.0	37.255.135.255	Iran Telecommunication Company Pjs
37.255.136.0	37.255.136.255	Iran Telecommunication Company Pjs
37.255.137.0	37.255.137.255	Iran Telecommunication Company Pjs
37.255.138.0	37.255.139.255	Iran Telecommunication Company Pjs
37.255.140.0	37.255.143.255	Iran Telecommunication Company Pjs
37.255.144.0	37.255.159.255	Iran Telecommunication Company Pjs
37.255.160.0	37.255.191.255	Iran Telecommunication Company Pjs
37.255.192.0	37.255.207.255	Iran Telecommunication Company Pjs
37.255.208.0	37.255.209.255	Iran Telecommunication Company Pjs
37.255.210.0	37.255.210.255	Iran Telecommunication Company Pjs
37.255.211.0	37.255.211.255	Iran Telecommunication Company Pjs
37.255.212.0	37.255.215.255	Iran Telecommunication Company Pjs
37.255.216.0	37.255.223.255	Iran Telecommunication Company Pjs
37.255.224.0	37.255.239.255	Iran Telecommunication Company Pjs
37.255.240.0	37.255.247.255	Iran Telecommunication Company Pjs
37.255.248.0	37.255.248.255	Iran Telecommunication Company Pjs
37.255.249.0	37.255.249.255	Iran Telecommunication Company Pjs
37.255.250.0	37.255.251.255	Iran Telecommunication Company Pjs
37.255.252.0	37.255.255.255	Iran Telecommunication Company Pjs
45.8.160.0	45.8.160.255	Farzanegan Pars Communications Company PJS
45.8.161.0	45.8.161.255	Farzanegan Pars Communications Company PJS
45.8.162.0	45.8.162.255	Farzanegan Pars Communications Company PJS
45.8.163.0	45.8.163.255	Farzanegan Pars Communications Company PJS
45.9.144.0	45.9.144.255	Pooya Parto Qeshm Cooperative Company
45.9.145.0	45.9.145.255	Pooya Parto Qeshm Cooperative Company
45.9.146.0	45.9.146.255	Pooya Parto Qeshm Cooperative Company
45.9.147.0	45.9.147.255	Pooya Parto Qeshm Cooperative Company
45.9.252.0	45.9.252.255	Sefroyek Pardaz Engineering PJSC
45.9.253.0	45.9.253.255	Sefroyek Pardaz Engineering PJSC
45.9.254.0	45.9.254.255	Sefroyek Pardaz Engineering PJSC
45.9.255.0	45.9.255.255	Simorq Tejarat Integrated Systems PJSC
45.11.184.0	45.11.184.255	Giti Secure Cloud LLC
45.11.185.0	45.11.185.255	Toosee Ertebatat Damavand
45.11.186.0	45.11.186.255	Iranian Server Processing Limited Liability Company
45.11.187.0	45.11.187.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
45.15.200.0	45.15.200.255	Sigma IT Infrastructures Development Co. (Ltd.)
45.15.201.0	45.15.201.255	Sigma IT Infrastructures Development Co. (Ltd.)
45.15.202.0	45.15.202.255	Sigma IT Infrastructures Development Co. (Ltd.)
45.15.203.0	45.15.203.255	Sigma IT Infrastructures Development Co. (Ltd.)
45.15.248.0	45.15.248.255	Asr-e Enteghal-e Dadeha Company (Private J.S.)
45.15.249.0	45.15.249.255	Asr-e Enteghal-e Dadeha Company (Private J.S.)
45.15.250.0	45.15.250.255	Asr-e Enteghal-e Dadeha Company (Private J.S.)
45.15.251.0	45.15.251.255	Asr-e Enteghal-e Dadeha Company (Private J.S.)
45.81.16.0	45.81.16.255	Amir Hosein Maaref
45.81.17.0	45.81.17.255	Amir Hosein Maaref
45.81.18.0	45.81.18.255	Nahor Hadish Design and Architecture Co. Ltd
45.81.19.0	45.81.19.255	Amir Hosein Maaref
45.82.136.0	45.82.137.255	Noyan Abr Arvan Co. ( Private Joint Stock)
45.82.138.0	45.82.139.255	Noyan Abr Arvan Co. ( Private Joint Stock)
45.83.12.0	45.83.15.255	Mobin Net Communication Company (Private Joint Stock)
45.84.156.0	45.84.156.255	Didehban Net Company PJS
45.84.157.0	45.84.157.255	Didehban Net Company PJS
45.84.158.0	45.84.159.255	Didehban Net Company PJS
45.84.248.0	45.84.251.255	Navaco Information Technology Company (P.J.S.)
45.86.4.0	45.86.4.255	Pooya Parto Qeshm Cooperative Company
45.86.5.0	45.86.5.255	Pooya Parto Qeshm Cooperative Company
45.86.6.0	45.86.6.255	Pooya Parto Qeshm Cooperative Company
45.86.7.0	45.86.7.255	Pooya Parto Qeshm Cooperative Company
45.86.87.0	45.86.87.255	Avini Cultural and Art Institute
45.86.196.0	45.86.196.255	Asr-e Enteghal-e Dadeha Company (Private J.S.)
45.86.197.0	45.86.197.255	Asr-e Enteghal-e Dadeha Company (Private J.S.)
45.86.198.0	45.86.198.255	Asr-e Enteghal-e Dadeha Company (Private J.S.)
45.86.199.0	45.86.199.255	Asr-e Enteghal-e Dadeha Company (Private J.S.)
45.87.4.0	45.87.5.255	Sefroyek Pardaz Engineering PJSC
45.87.6.0	45.87.7.255	Asiatech Data Transmission Company
45.89.136.0	45.89.136.255	Hamrah Pardaz Zarin PJS
45.89.137.0	45.89.137.255	Hamrah Pardaz Zarin PJS
45.89.138.0	45.89.138.255	Hamrah Pardaz Zarin PJS
45.89.139.0	45.89.139.255	Hamrah Pardaz Zarin PJS
45.89.200.0	45.89.200.255	Alibaba Travel Company (LTD)
45.89.201.0	45.89.201.255	Alibaba Travel Company (LTD)
45.89.202.0	45.89.202.255	Alibaba Travel Company (LTD)
45.89.203.0	45.89.203.255	Alibaba Travel Company (LTD)
45.89.221.0	45.89.221.255	Seyed Ali Moosavi
45.89.222.0	45.89.222.255	Negah Roshan Pars Company (PJS)
45.89.236.0	45.89.236.255	Faraso Samaneh Pasargad Ltd.
45.89.237.0	45.89.237.255	Faraso Samaneh Pasargad Ltd.
45.89.238.0	45.89.238.255	Faraso Samaneh Pasargad Ltd.
45.89.239.0	45.89.239.255	Faraso Samaneh Pasargad Ltd.
45.90.72.0	45.90.72.255	Mizban Dade Pasargad LLC
45.90.73.0	45.90.73.255	Mizban Dade Pasargad LLC
45.90.74.0	45.90.74.255	Mizban Dade Pasargad LLC
45.90.75.0	45.90.75.255	Mizban Dade Pasargad LLC
45.91.152.0	45.91.152.255	Shahr Bank PJSC
45.91.153.0	45.91.153.255	Shahr Bank PJSC
45.91.154.0	45.91.154.255	Shahr Bank PJSC
45.91.155.0	45.91.155.255	Shahr Bank PJSC
45.92.92.0	45.92.92.255	Noavaran Shabakeh Sabz Mehregan (Ltd.)
45.92.93.0	45.92.93.255	Noavaran Shabakeh Sabz Mehregan (Ltd.)
45.92.94.0	45.92.94.255	Noavaran Shabakeh Sabz Mehregan (Ltd.)
45.92.95.0	45.92.95.255	Noavaran Shabakeh Sabz Mehregan (Ltd.)
45.94.212.0	45.94.212.255	Hormozgan University of Medical Sciences
45.94.213.0	45.94.213.255	Vandad Vira Hooman LLC
45.94.214.0	45.94.214.255	Vandad Vira Hooman LLC
45.94.215.0	45.94.215.255	Vandad Vira Hooman LLC
45.94.252.0	45.94.252.255	Sindad Network Technology PJSC
45.94.253.0	45.94.253.255	Tose'e Vistaye Mohadelat Nadin Company
45.94.254.0	45.94.254.255	Sindad Network Technology PJSC
45.94.255.0	45.94.255.255	Sindad Network Technology PJSC
45.95.88.0	45.95.91.255	Mobin Net Communication Company (Private Joint Stock)
45.128.140.0	45.128.143.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
45.129.36.0	45.129.36.255	Dade Pardazi Mobinhost Co Ltd
45.129.37.0	45.129.37.255	Dade Pardazi Mobinhost Co Ltd
45.129.38.0	45.129.38.255	Dade Pardazi Mobinhost Co Ltd
45.129.39.0	45.129.39.255	Dade Pardazi Mobinhost Co Ltd
45.129.116.0	45.129.116.255	Pasargad Bank PJSC
45.129.117.0	45.129.117.255	Pasargad Bank PJSC
45.129.118.0	45.129.118.255	Pasargad Bank PJSC
45.129.119.0	45.129.119.255	Pasargad Bank PJSC
45.132.32.0	45.132.32.255	Didehban Net Company PJS
45.132.168.0	45.132.171.255	Asr-e Enteghal-e Dadeha Company (Private J.S.)
45.132.172.0	45.132.172.255	Pardazeshgar Ray Azma Co. Ltd.
45.132.173.0	45.132.173.255	Pardazeshgar Ray Azma Co. Ltd.
45.132.174.0	45.132.174.255	Pardazeshgar Ray Azma Co. Ltd.
45.132.175.0	45.132.175.255	Pardazeshgar Ray Azma Co. Ltd.
45.134.97.0	45.134.97.255	Tida Data Processing Ltd
45.134.99.0	45.134.99.255	Didehban Net Company PJS
45.135.240.0	45.135.240.255	Sabz Pardazesh System Ghadir Co. (P.J.S.)
45.135.241.0	45.135.241.255	Asre Pardazeshe Ettelaate Amin Institute
45.135.242.0	45.135.242.255	Faraso Samaneh Pasargad Ltd.
45.135.243.0	45.135.243.255	Faraso Samaneh Pasargad Ltd.
45.137.16.0	45.137.16.255	Pishgam Rayan Dadeh Pajooh PJSC
45.137.17.0	45.137.17.255	Pishgam Rayan Dadeh Pajooh PJSC
45.137.19.0	45.137.19.255	Pishgam Rayan Dadeh Pajooh PJSC
45.138.132.0	45.138.132.255	Noavaran Shabakeh Sabz Mehregan (Ltd.)
45.138.133.0	45.138.133.255	Noavaran Shabakeh Sabz Mehregan (Ltd.)
45.138.134.0	45.138.134.255	Noavaran Shabakeh Sabz Mehregan (Ltd.)
45.138.135.0	45.138.135.255	Noavaran Shabakeh Sabz Mehregan (Ltd.)
45.139.8.0	45.139.9.255	Patron Technology Persia Ltd
45.139.10.0	45.139.10.255	Vandad Vira Hooman LLC
45.139.11.0	45.139.11.255	Vandad Vira Hooman LLC
45.140.28.0	45.140.31.255	Parvaresh Dadeha Co. Private Joint Stock
45.140.224.0	45.140.227.255	Sefroyek Pardaz Engineering PJSC
45.140.228.0	45.140.231.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
45.142.188.0	45.142.188.255	Cooperative Afra ertebatat-e-sabet-e Rasa Co
45.142.189.0	45.142.189.255	Cooperative Afra ertebatat-e-sabet-e Rasa Co
45.142.190.0	45.142.190.255	Cooperative Afra ertebatat-e-sabet-e Rasa Co
45.142.191.0	45.142.191.255	Cooperative Afra ertebatat-e-sabet-e Rasa Co
45.143.252.0	45.143.255.255	Mobin Net Communication Company (Private Joint Stock)
45.144.16.0	45.144.16.255	Foojan Cloud Computing Co. PJS
45.144.17.0	45.144.17.255	Faragostar Shargh Communication Company PJS
45.144.18.0	45.144.18.255	Fanavaran Asak Mahan Asia Ltd.
45.144.19.0	45.144.19.255	Foojan Cloud Computing Co. PJS
45.144.124.0	45.144.127.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
45.146.240.0	45.146.240.255	Fannavari Etelaate Samane Kavoshgar Ide Ltd.
45.147.76.0	45.147.76.255	Gostaresh Pardazesh Dana Negar Co.(PJSC)
45.147.77.0	45.147.77.255	Gostaresh Pardazesh Dana Negar Co.(PJSC)
45.147.78.0	45.147.78.255	Gostaresh Pardazesh Dana Negar Co.(PJSC)
45.147.79.0	45.147.79.255	Gostaresh Pardazesh Dana Negar Co.(PJSC)
45.148.248.0	45.148.248.255	Pooya Parto Qeshm Cooperative Company
45.148.249.0	45.148.249.255	Pooya Parto Qeshm Cooperative Company
45.148.250.0	45.148.250.255	Pooya Parto Qeshm Cooperative Company
45.148.251.0	45.148.251.255	Pooya Parto Qeshm Cooperative Company
45.149.76.0	45.149.76.255	Vandad Vira Hooman LLC
45.149.77.0	45.149.77.255	Vandad Vira Hooman LLC
45.149.78.0	45.149.78.255	Vandad Vira Hooman LLC
45.149.79.0	45.149.79.255	Vandad Vira Hooman LLC
45.150.88.0	45.150.91.255	Iran Fara Bourse Co. (Public Joint Stock)
45.150.150.0	45.150.150.255	Sindad Network Technology PJSC
45.155.192.0	45.155.192.255	Hesabgar Pardaz Gharb PJSC
45.155.193.0	45.155.193.255	Hesabgar Pardaz Gharb PJSC
45.155.194.0	45.155.194.255	Hesabgar Pardaz Gharb PJSC
45.155.195.0	45.155.195.255	Hesabgar Pardaz Gharb PJSC
45.156.116.0	45.156.117.255	Pishgam Rayan Dadeh Pajooh PJSC
45.156.118.0	45.156.118.255	Pishgam Rayan Dadeh Pajooh PJSC
45.156.119.0	45.156.119.255	Pishgam Rayan Dadeh Pajooh PJSC
45.156.180.0	45.156.183.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
45.156.184.0	45.156.184.255	Pars Shabakeh Azarakhsh LLC
45.156.185.0	45.156.185.255	Pars Shabakeh Azarakhsh LLC
45.156.186.0	45.156.186.255	Pars Shabakeh Azarakhsh LLC
45.156.187.0	45.156.187.255	Pars Shabakeh Azarakhsh LLC
45.156.192.0	45.156.195.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
45.156.196.0	45.156.196.255	Asr-e Enteghal-e Dadeha Company (Private J.S.)
45.156.197.0	45.156.197.255	Asr-E Danesh Afzar Company (Private J.S.)
45.156.198.0	45.156.198.255	Asr-e Enteghal-e Dadeha Company (Private J.S.)
45.156.199.0	45.156.199.255	Asr-E Danesh Afzar Company (Private J.S.)
45.156.200.0	45.156.200.255	Tadbir Pardaz IT Group LTD
45.156.201.0	45.156.201.255	Tadbir Pardaz IT Group LTD
45.156.202.0	45.156.202.255	Tadbir Pardaz IT Group LTD
45.156.203.0	45.156.203.255	Tadbir Pardaz IT Group LTD
45.157.244.0	45.157.244.255	Fars News Agency Cultural Arts Institute
45.157.245.0	45.157.245.255	Fars News Agency Cultural Arts Institute
45.157.246.0	45.157.247.255	Fars News Agency Cultural Arts Institute
45.158.120.0	45.158.120.255	University of Isfahan
45.158.121.0	45.158.121.255	University of Isfahan
45.158.122.0	45.158.123.255	University of Isfahan
45.159.112.0	45.159.112.255	Green Web Samaneh Novin PJSC
45.159.113.0	45.159.113.255	Green Web Samaneh Novin PJSC
45.159.114.0	45.159.114.255	Green Web Samaneh Novin PJSC
45.159.115.0	45.159.115.255	Green Web Samaneh Novin PJSC
45.159.148.0	45.159.148.255	Hossein Zangooei
45.159.149.0	45.159.149.255	Vandad Vira Hooman LLC
45.159.150.0	45.159.150.255	Vandad Vira Hooman LLC
45.159.151.0	45.159.151.255	Fannavari Etelaate Samane Kavoshgar Ide Ltd.
45.159.196.0	45.159.196.255	Sindad Network Technology PJSC
45.159.197.0	45.159.197.255	Sindad Network Technology PJSC
45.159.198.0	45.159.198.255	Sindad Network Technology PJSC
45.159.199.0	45.159.199.255	Sindad Network Technology PJSC
45.248.164.0	45.248.167.255	Mobin Net Communication Company (Private Joint Stock)
46.18.248.0	46.18.248.255	Karafarin Bank Plc.
46.18.249.0	46.18.249.255	Karafarin Bank Plc.
46.18.250.0	46.18.250.255	Karafarin Bank Plc.
46.18.251.0	46.18.251.255	Karafarin Bank Plc.
46.18.252.0	46.18.255.255	Karafarin Bank Plc.
46.28.72.0	46.28.72.255	Sefroyek Pardaz Engineering PJSC
46.28.73.0	46.28.73.255	Sefroyek Pardaz Engineering PJSC
46.28.74.0	46.28.74.255	Sefroyek Pardaz Engineering PJSC
46.28.75.0	46.28.75.255	Sefroyek Pardaz Engineering PJSC
46.28.76.0	46.28.76.255	Sefroyek Pardaz Engineering PJSC
46.28.77.0	46.28.77.255	Sefroyek Pardaz Engineering PJSC
46.28.78.0	46.28.78.255	Tam Iran Khodro Company
46.28.79.0	46.28.79.255	Sefroyek Pardaz Engineering PJSC
46.29.32.0	46.29.32.255	Mahdi Dehestani
46.29.33.0	46.29.33.255	Maavaraaye Abaade Raika PJSC
46.29.34.0	46.29.34.255	Web Dadeh Paydar Co (Ltd)
46.32.0.0	46.32.0.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
46.32.1.0	46.32.1.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
46.32.2.0	46.32.2.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
46.32.3.0	46.32.3.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
46.32.4.0	46.32.4.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
46.32.5.0	46.32.5.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
46.32.6.0	46.32.6.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
46.32.7.0	46.32.7.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
46.32.8.0	46.32.8.255	Khatam Academic Institute
46.32.9.0	46.32.9.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
46.32.10.0	46.32.11.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
46.32.12.0	46.32.12.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
46.32.13.0	46.32.13.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
46.32.14.0	46.32.14.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
46.32.15.0	46.32.15.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
46.32.16.0	46.32.23.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
46.32.24.0	46.32.24.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
46.32.25.0	46.32.25.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
46.32.26.0	46.32.27.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
46.32.28.0	46.32.29.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
46.32.30.0	46.32.30.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
46.32.31.0	46.32.31.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
46.34.96.0	46.34.103.255	Mellat Insurance Public Joint Stock Company
46.34.104.0	46.34.111.255	Mellat Insurance Public Joint Stock Company
46.34.112.0	46.34.127.255	Mellat Insurance Public Joint Stock Company
46.34.160.0	46.34.160.255	Cooperative Afra ertebatat-e-sabet-e Rasa Co
46.34.161.0	46.34.161.255	Cooperative Afra ertebatat-e-sabet-e Rasa Co
46.34.162.0	46.34.162.255	Cooperative Afra ertebatat-e-sabet-e Rasa Co
46.34.163.0	46.34.163.255	Tose'e Shabakeh Azad Ltd
46.34.164.0	46.34.167.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
46.34.168.0	46.34.175.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
46.34.176.0	46.34.177.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
46.34.178.0	46.34.179.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
46.34.180.0	46.34.183.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
46.34.184.0	46.34.187.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
46.34.188.0	46.34.191.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
46.36.96.0	46.36.96.255	Pirooz Leen LLC
46.36.97.0	46.36.97.255	Mobin Net Communication Company (Private Joint Stock)
46.36.98.0	46.36.98.255	Mobin Net Communication Company (Private Joint Stock)
46.36.99.0	46.36.99.255	AMN Gostare Hasin Co
46.36.100.0	46.36.100.255	Mobin Net Communication Company (Private Joint Stock)
46.36.101.0	46.36.101.255	Mobin Net Communication Company (Private Joint Stock)
46.36.102.0	46.36.102.255	Mobin Net Communication Company (Private Joint Stock)
46.36.103.0	46.36.103.255	Mobin Net Communication Company (Private Joint Stock)
46.36.104.0	46.36.104.255	Pirooz Leen LLC
46.36.105.0	46.36.105.255	Mobin Net Communication Company (Private Joint Stock)
46.36.106.0	46.36.106.255	Mobin Net Communication Company (Private Joint Stock)
46.36.107.0	46.36.107.255	Pirooz Leen LLC
46.36.108.0	46.36.108.255	Pirooz Leen LLC
46.36.109.0	46.36.109.255	Pirooz Leen LLC
46.36.110.0	46.36.110.255	Pirooz Leen LLC
46.36.111.0	46.36.111.255	Sarmayeh & Danesh Brokerage Co.
46.38.129.0	46.38.129.255	Institute for Research in Fundamental Sciences
46.38.131.0	46.38.131.255	Institute for Research in Fundamental Sciences
46.38.132.0	46.38.132.255	Modiriat Saderat Ozhan Mashregh Pars Company PJS
46.38.133.0	46.38.133.255	Modiriat Saderat Ozhan Mashregh Pars Company PJS
46.38.134.0	46.38.134.255	Modiriat Saderat Ozhan Mashregh Pars Company PJS
46.38.135.0	46.38.135.255	Modiriat Saderat Ozhan Mashregh Pars Company PJS
46.38.136.0	46.38.137.255	Farhang Azma Communications Company LTD
46.38.138.0	46.38.138.255	Giti Secure Cloud LLC
46.38.139.0	46.38.139.255	Farhang Azma Communications Company LTD
46.38.140.0	46.38.140.255	Farhang Azma Communications Company LTD
46.38.142.0	46.38.142.255	Farhang Azma Communications Company LTD
46.38.143.0	46.38.143.255	Giti Secure Cloud LLC
46.38.144.0	46.38.144.255	Afranet
46.38.145.0	46.38.145.255	Farhang Azma Communications Company LTD
46.38.146.0	46.38.146.255	Peyman Ertebatat Pouya Company (Ltd)
46.38.147.0	46.38.147.255	Farhang Azma Communications Company LTD
46.38.148.0	46.38.148.255	Parmin Cloud Computing LLC
46.38.149.0	46.38.149.255	Parmin Cloud Computing LLC
46.38.150.0	46.38.151.255	Farhang Azma Communications Company LTD
46.38.154.0	46.38.154.255	Farhang Azma Communications Company LTD
46.38.158.0	46.38.158.255	Farhang Azma Communications Company LTD
46.38.159.0	46.38.159.255	Iranian Electronic Counter Anzali Free Zone(Private Joint Stock Company)
46.41.192.0	46.41.255.255	Parsan Lin Co. PJS
46.51.0.0	46.51.0.255	Mobile Communication Company of Iran Plc
46.51.1.0	46.51.1.255	Mobile Communication Company of Iran Plc
46.51.2.0	46.51.3.255	Mobile Communication Company of Iran Plc
46.51.4.0	46.51.7.255	Mobile Communication Company of Iran Plc
46.51.8.0	46.51.15.255	Mobile Communication Company of Iran Plc
46.51.16.0	46.51.31.255	Mobile Communication Company of Iran Plc
46.51.32.0	46.51.35.255	Mobile Communication Company of Iran Plc
46.51.36.0	46.51.36.255	Mobile Communication Company of Iran Plc
46.51.37.0	46.51.37.255	Mobile Communication Company of Iran Plc
46.51.38.0	46.51.38.255	Mobile Communication Company of Iran Plc
46.51.39.0	46.51.39.255	Mobile Communication Company of Iran Plc
46.51.40.0	46.51.47.255	Mobile Communication Company of Iran Plc
46.51.48.0	46.51.63.255	Mobile Communication Company of Iran Plc
46.51.64.0	46.51.79.255	Mobile Communication Company of Iran Plc
46.51.80.0	46.51.95.255	Mobile Communication Company of Iran Plc
46.51.96.0	46.51.111.255	Mobile Communication Company of Iran Plc
46.51.112.0	46.51.127.255	Mobile Communication Company of Iran Plc
46.100.0.0	46.100.3.255	Iran Telecommunication Company Pjs
46.100.4.0	46.100.4.255	Iran Telecommunication Company Pjs
46.100.5.0	46.100.5.255	Iran Telecommunication Company Pjs
46.100.6.0	46.100.7.255	Iran Telecommunication Company Pjs
46.100.8.0	46.100.15.255	Iran Telecommunication Company Pjs
46.100.16.0	46.100.31.255	Iran Telecommunication Company Pjs
46.100.32.0	46.100.47.255	Iran Telecommunication Company Pjs
46.100.48.0	46.100.55.255	Iran Telecommunication Company Pjs
46.100.56.0	46.100.57.255	Iran Telecommunication Company Pjs
46.100.58.0	46.100.58.255	Iran Telecommunication Company Pjs
46.100.59.0	46.100.59.255	Iran Telecommunication Company Pjs
46.100.60.0	46.100.60.255	Iran Telecommunication Company Pjs
46.100.61.0	46.100.61.255	Iran Telecommunication Company Pjs
46.100.62.0	46.100.63.255	Iran Telecommunication Company Pjs
46.100.64.0	46.100.67.255	Iran Telecommunication Company Pjs
46.100.68.0	46.100.71.255	Iran Telecommunication Company Pjs
46.100.72.0	46.100.75.255	Iran Telecommunication Company Pjs
46.100.76.0	46.100.79.255	Iran Telecommunication Company Pjs
46.100.80.0	46.100.83.255	Iran Telecommunication Company Pjs
46.100.84.0	46.100.85.255	Iran Telecommunication Company Pjs
46.100.86.0	46.100.86.255	Iran Telecommunication Company Pjs
46.100.87.0	46.100.87.255	Iran Telecommunication Company Pjs
46.100.88.0	46.100.95.255	Iran Telecommunication Company Pjs
46.100.96.0	46.100.97.255	Iran Telecommunication Company Pjs
46.100.98.0	46.100.98.255	Iran Telecommunication Company Pjs
46.100.99.0	46.100.99.255	Iran Telecommunication Company Pjs
46.100.100.0	46.100.101.255	Iran Telecommunication Company Pjs
46.100.102.0	46.100.103.255	Iran Telecommunication Company Pjs
46.100.104.0	46.100.105.255	Iran Telecommunication Company Pjs
46.100.106.0	46.100.107.255	Iran Telecommunication Company Pjs
46.100.108.0	46.100.111.255	Iran Telecommunication Company Pjs
46.100.112.0	46.100.113.255	Iran Telecommunication Company Pjs
46.100.114.0	46.100.114.255	Iran Telecommunication Company Pjs
46.100.115.0	46.100.115.255	Iran Telecommunication Company Pjs
46.100.116.0	46.100.117.255	Iran Telecommunication Company Pjs
46.100.118.0	46.100.119.255	Iran Telecommunication Company Pjs
46.100.120.0	46.100.127.255	Iran Telecommunication Company Pjs
46.100.128.0	46.100.131.255	Iran Telecommunication Company Pjs
46.100.132.0	46.100.133.255	Iran Telecommunication Company Pjs
46.100.134.0	46.100.135.255	Iran Telecommunication Company Pjs
46.100.136.0	46.100.139.255	Iran Telecommunication Company Pjs
46.100.140.0	46.100.140.255	Iran Telecommunication Company Pjs
46.100.141.0	46.100.141.255	Iran Telecommunication Company Pjs
46.100.142.0	46.100.143.255	Iran Telecommunication Company Pjs
46.100.144.0	46.100.151.255	Iran Telecommunication Company Pjs
46.100.152.0	46.100.155.255	Iran Telecommunication Company Pjs
46.100.156.0	46.100.159.255	Iran Telecommunication Company Pjs
46.100.160.0	46.100.160.255	Iran Telecommunication Company Pjs
46.100.161.0	46.100.161.255	Iran Telecommunication Company Pjs
46.100.162.0	46.100.163.255	Iran Telecommunication Company Pjs
46.100.164.0	46.100.167.255	Iran Telecommunication Company Pjs
46.100.168.0	46.100.171.255	Iran Telecommunication Company Pjs
46.100.172.0	46.100.175.255	Iran Telecommunication Company Pjs
46.100.176.0	46.100.179.255	Iran Telecommunication Company Pjs
46.100.180.0	46.100.183.255	Iran Telecommunication Company Pjs
46.100.184.0	46.100.187.255	Iran Telecommunication Company Pjs
46.100.188.0	46.100.191.255	Iran Telecommunication Company Pjs
46.100.192.0	46.100.207.255	Iran Telecommunication Company Pjs
46.100.208.0	46.100.215.255	Iran Telecommunication Company Pjs
46.100.216.0	46.100.223.255	Iran Telecommunication Company Pjs
46.100.224.0	46.100.227.255	Iran Telecommunication Company Pjs
46.100.228.0	46.100.231.255	Iran Telecommunication Company Pjs
46.100.232.0	46.100.233.255	Iran Telecommunication Company Pjs
46.100.234.0	46.100.235.255	Iran Telecommunication Company Pjs
46.100.236.0	46.100.239.255	Iran Telecommunication Company Pjs
46.100.240.0	46.100.243.255	Iran Telecommunication Company Pjs
46.100.244.0	46.100.247.255	Iran Telecommunication Company Pjs
46.100.248.0	46.100.251.255	Iran Telecommunication Company Pjs
46.100.252.0	46.100.253.255	Iran Telecommunication Company Pjs
46.100.254.0	46.100.255.255	Iran Telecommunication Company Pjs
46.102.120.0	46.102.127.255	Iran Telecommunication Company Pjs
46.102.128.0	46.102.128.255	Afranet
46.102.129.0	46.102.129.255	Afranet
46.102.130.0	46.102.130.255	Afranet
46.102.131.0	46.102.131.255	Afranet
46.102.132.0	46.102.132.255	Afranet
46.102.133.0	46.102.133.255	Afranet
46.102.134.0	46.102.134.255	Afranet
46.102.135.0	46.102.135.255	Afranet
46.102.136.0	46.102.136.255	Afranet
46.102.137.0	46.102.137.255	Afranet
46.102.138.0	46.102.138.255	Afranet
46.102.139.0	46.102.139.255	Afranet
46.102.140.0	46.102.140.255	Afranet
46.102.141.0	46.102.141.255	Afranet
46.102.142.0	46.102.142.255	Afranet
46.102.143.0	46.102.143.255	Afranet
46.102.184.0	46.102.185.255	Iran Telecommunication Company Pjs
46.102.186.0	46.102.186.255	Iran Telecommunication Company Pjs
46.102.187.0	46.102.187.255	Iran Telecommunication Company Pjs
46.143.0.0	46.143.3.255	Asiatech Data Transmission Company
46.143.4.0	46.143.7.255	Asiatech Data Transmission Company
46.143.8.0	46.143.11.255	Asiatech Data Transmission Company
46.143.12.0	46.143.15.255	Asiatech Data Transmission Company
46.143.16.0	46.143.19.255	Asiatech Data Transmission Company
46.143.20.0	46.143.23.255	Asiatech Data Transmission Company
46.143.24.0	46.143.27.255	Asiatech Data Transmission Company
46.143.28.0	46.143.31.255	Asiatech Data Transmission Company
46.143.32.0	46.143.35.255	Asiatech Data Transmission Company
46.143.36.0	46.143.39.255	Asiatech Data Transmission Company
46.143.40.0	46.143.43.255	Asiatech Data Transmission Company
46.143.44.0	46.143.47.255	Asiatech Data Transmission Company
46.143.48.0	46.143.51.255	Asiatech Data Transmission Company
46.143.52.0	46.143.55.255	Asiatech Data Transmission Company
46.143.56.0	46.143.59.255	Asiatech Data Transmission Company
46.143.60.0	46.143.63.255	Asiatech Data Transmission Company
46.143.64.0	46.143.67.255	Asiatech Data Transmission Company
46.143.68.0	46.143.71.255	Asiatech Data Transmission Company
46.143.72.0	46.143.75.255	Asiatech Data Transmission Company
46.143.76.0	46.143.79.255	Asiatech Data Transmission Company
46.143.80.0	46.143.83.255	Asiatech Data Transmission Company
46.143.84.0	46.143.87.255	Asiatech Data Transmission Company
46.143.88.0	46.143.91.255	Asiatech Data Transmission Company
46.143.92.0	46.143.95.255	Asiatech Data Transmission Company
46.143.96.0	46.143.99.255	Asiatech Data Transmission Company
46.143.100.0	46.143.103.255	Asiatech Data Transmission Company
46.143.104.0	46.143.107.255	Asiatech Data Transmission Company
46.143.108.0	46.143.111.255	Asiatech Data Transmission Company
46.143.112.0	46.143.115.255	Asiatech Data Transmission Company
46.143.116.0	46.143.119.255	Asiatech Data Transmission Company
46.143.120.0	46.143.123.255	Asiatech Data Transmission Company
46.143.124.0	46.143.127.255	Asiatech Data Transmission Company
46.143.192.0	46.143.223.255	Mihan Communication Systems Co. Ltd
46.143.224.0	46.143.239.255	Mihan Communication Systems Co. Ltd
46.143.240.0	46.143.243.255	Mihan Communication Systems Co. Ltd
46.143.244.0	46.143.244.255	Sepahan Mapna Equipment Engineering and Manufacturing PJSC
46.143.245.0	46.143.245.255	Sepahan Mapna Equipment Engineering and Manufacturing PJSC
46.143.246.0	46.143.247.255	Mihan Communication Systems Co. Ltd
46.143.248.0	46.143.255.255	Mihan Communication Systems Co. Ltd
46.148.32.0	46.148.32.255	Respina Networks & Beyond PJSC
46.148.33.0	46.148.33.255	Abr Gostar Arianet LLC
46.148.34.0	46.148.34.255	Part Payam Paya LLC
46.148.35.0	46.148.35.255	Pooya Parto Qeshm Cooperative Company
46.148.36.0	46.148.36.255	Pooya Parto Qeshm Cooperative Company
46.148.37.0	46.148.37.255	Abr Gostar Arianet LLC
46.148.38.0	46.148.38.255	Pooya Parto Qeshm Cooperative Company
46.148.39.0	46.148.39.255	Hormoz IT & Network Waves Connection Co. (PJS)
46.148.40.0	46.148.40.255	Pooya Parto Qeshm Cooperative Company
46.148.41.0	46.148.41.255	Web Gostaran Bandar Company (PJS)
46.148.42.0	46.148.42.255	Part Payam Paya LLC
46.148.43.0	46.148.43.255	Pooya Parto Qeshm Cooperative Company
46.148.44.0	46.148.44.255	Pooya Parto Qeshm Cooperative Company
46.148.45.0	46.148.45.255	Pooya Parto Qeshm Cooperative Company
46.148.46.0	46.148.46.255	Pooya Parto Qeshm Cooperative Company
46.148.47.0	46.148.47.255	Part Payam Paya LLC
46.164.64.0	46.164.64.255	Mobile Communication Company of Iran Plc
46.164.65.0	46.164.65.255	Mobile Communication Company of Iran Plc
46.164.66.0	46.164.67.255	Mobile Communication Company of Iran Plc
46.164.68.0	46.164.71.255	Mobile Communication Company of Iran Plc
46.164.72.0	46.164.79.255	Mobile Communication Company of Iran Plc
46.164.80.0	46.164.95.255	Mobile Communication Company of Iran Plc
46.164.96.0	46.164.111.255	Mobile Communication Company of Iran Plc
46.164.112.0	46.164.127.255	Mobile Communication Company of Iran Plc
46.167.128.0	46.167.159.255	Parsan Lin Co. PJS
46.182.32.0	46.182.39.255	Iran Telecommunication Company Pjs
46.209.0.0	46.209.1.255	Respina Networks & Beyond PJSC
46.209.2.0	46.209.2.255	Respina Networks & Beyond PJSC
46.209.3.0	46.209.3.255	Respina Networks & Beyond PJSC
46.209.4.0	46.209.4.255	Respina Networks & Beyond PJSC
46.209.5.0	46.209.5.255	Respina Networks & Beyond PJSC
46.209.6.0	46.209.6.255	Respina Networks & Beyond PJSC
46.209.7.0	46.209.7.255	Respina Networks & Beyond PJSC
46.209.8.0	46.209.8.255	Respina Networks & Beyond PJSC
46.209.9.0	46.209.9.255	Respina Networks & Beyond PJSC
46.209.10.0	46.209.10.255	Respina Networks & Beyond PJSC
46.209.11.0	46.209.11.255	Respina Networks & Beyond PJSC
46.209.12.0	46.209.13.255	Respina Networks & Beyond PJSC
46.209.14.0	46.209.14.255	Respina Networks & Beyond PJSC
46.209.15.0	46.209.15.255	Respina Networks & Beyond PJSC
46.209.16.0	46.209.16.255	Respina Networks & Beyond PJSC
46.209.17.0	46.209.17.255	Respina Networks & Beyond PJSC
46.209.18.0	46.209.18.255	Respina Networks & Beyond PJSC
46.209.19.0	46.209.19.255	Respina Networks & Beyond PJSC
46.209.20.0	46.209.20.255	Respina Networks & Beyond PJSC
46.209.21.0	46.209.21.255	Respina Networks & Beyond PJSC
46.209.22.0	46.209.22.255	Respina Networks & Beyond PJSC
46.209.23.0	46.209.23.255	Respina Networks & Beyond PJSC
46.209.24.0	46.209.24.255	Respina Networks & Beyond PJSC
46.209.25.0	46.209.25.255	Respina Networks & Beyond PJSC
46.209.26.0	46.209.26.255	Respina Networks & Beyond PJSC
46.209.27.0	46.209.27.255	Respina Networks & Beyond PJSC
46.209.28.0	46.209.28.255	Respina Networks & Beyond PJSC
46.209.29.0	46.209.29.255	Respina Networks & Beyond PJSC
46.209.30.0	46.209.30.255	Respina Networks & Beyond PJSC
46.209.31.0	46.209.31.255	Respina Networks & Beyond PJSC
46.209.32.0	46.209.32.255	Shiraz University
46.209.33.0	46.209.33.255	Respina Networks & Beyond PJSC
46.209.34.0	46.209.34.255	Shiraz University
46.209.35.0	46.209.35.255	Shiraz University
46.209.36.0	46.209.36.255	Shiraz University
46.209.37.0	46.209.37.255	Shiraz University
46.209.38.0	46.209.38.255	Shiraz University
46.209.39.0	46.209.39.255	Shiraz University
46.209.40.0	46.209.43.255	Respina Networks & Beyond PJSC
46.209.44.0	46.209.44.255	Respina Networks & Beyond PJSC
46.209.45.0	46.209.45.255	Respina Networks & Beyond PJSC
46.209.46.0	46.209.46.255	Respina Networks & Beyond PJSC
46.209.47.0	46.209.47.255	Respina Networks & Beyond PJSC
46.209.48.0	46.209.48.255	Respina Networks & Beyond PJSC
46.209.49.0	46.209.49.255	Respina Networks & Beyond PJSC
46.209.50.0	46.209.51.255	Respina Networks & Beyond PJSC
46.209.52.0	46.209.55.255	Respina Networks & Beyond PJSC
46.209.56.0	46.209.57.255	Respina Networks & Beyond PJSC
46.209.58.0	46.209.58.255	Respina Networks & Beyond PJSC
46.209.59.0	46.209.59.255	Respina Networks & Beyond PJSC
46.209.60.0	46.209.60.255	Respina Networks & Beyond PJSC
46.209.61.0	46.209.61.255	Respina Networks & Beyond PJSC
46.209.62.0	46.209.62.255	Respina Networks & Beyond PJSC
46.209.63.0	46.209.63.255	Respina Networks & Beyond PJSC
46.209.64.0	46.209.64.255	Respina Networks & Beyond PJSC
46.209.65.0	46.209.65.255	Respina Networks & Beyond PJSC
46.209.66.0	46.209.66.255	Respina Networks & Beyond PJSC
46.209.67.0	46.209.67.255	Respina Networks & Beyond PJSC
46.209.68.0	46.209.68.255	Respina Networks & Beyond PJSC
46.209.69.0	46.209.69.255	Respina Networks & Beyond PJSC
46.209.70.0	46.209.70.255	Respina Networks & Beyond PJSC
46.209.71.0	46.209.71.255	Respina Networks & Beyond PJSC
46.209.72.0	46.209.72.255	Respina Networks & Beyond PJSC
46.209.73.0	46.209.73.255	Respina Networks & Beyond PJSC
46.209.74.0	46.209.74.255	Respina Networks & Beyond PJSC
46.209.75.0	46.209.75.255	Respina Networks & Beyond PJSC
46.209.76.0	46.209.76.255	Respina Networks & Beyond PJSC
46.209.77.0	46.209.77.255	Respina Networks & Beyond PJSC
46.209.78.0	46.209.78.255	Respina Networks & Beyond PJSC
46.209.79.0	46.209.79.255	Respina Networks & Beyond PJSC
46.209.80.0	46.209.87.255	Respina Networks & Beyond PJSC
46.209.88.0	46.209.88.255	Respina Networks & Beyond PJSC
46.209.89.0	46.209.89.255	Respina Networks & Beyond PJSC
46.209.90.0	46.209.91.255	Respina Networks & Beyond PJSC
46.209.92.0	46.209.92.255	Respina Networks & Beyond PJSC
46.209.93.0	46.209.93.255	Respina Networks & Beyond PJSC
46.209.94.0	46.209.94.255	Respina Networks & Beyond PJSC
46.209.95.0	46.209.95.255	Respina Networks & Beyond PJSC
46.209.96.0	46.209.96.255	Respina Networks & Beyond PJSC
46.209.97.0	46.209.97.255	Respina Networks & Beyond PJSC
46.209.98.0	46.209.98.255	Respina Networks & Beyond PJSC
46.209.99.0	46.209.99.255	Respina Networks & Beyond PJSC
46.209.100.0	46.209.100.255	Respina Networks & Beyond PJSC
46.209.101.0	46.209.101.255	Respina Networks & Beyond PJSC
46.209.102.0	46.209.102.255	Respina Networks & Beyond PJSC
46.209.103.0	46.209.103.255	Respina Networks & Beyond PJSC
46.209.104.0	46.209.104.255	Respina Networks & Beyond PJSC
46.209.105.0	46.209.105.255	Respina Networks & Beyond PJSC
46.209.106.0	46.209.106.255	Respina Networks & Beyond PJSC
46.209.107.0	46.209.107.255	Respina Networks & Beyond PJSC
46.209.108.0	46.209.111.255	Respina Networks & Beyond PJSC
46.209.112.0	46.209.112.255	Respina Networks & Beyond PJSC
46.209.113.0	46.209.113.255	Respina Networks & Beyond PJSC
46.209.114.0	46.209.114.255	Respina Networks & Beyond PJSC
46.209.115.0	46.209.115.255	Respina Networks & Beyond PJSC
46.209.116.0	46.209.116.255	Respina Networks & Beyond PJSC
46.209.117.0	46.209.117.255	Respina Networks & Beyond PJSC
46.209.118.0	46.209.118.255	Respina Networks & Beyond PJSC
46.209.119.0	46.209.119.255	Respina Networks & Beyond PJSC
46.209.120.0	46.209.120.255	Respina Networks & Beyond PJSC
46.209.121.0	46.209.121.255	Respina Networks & Beyond PJSC
46.209.122.0	46.209.122.255	Respina Networks & Beyond PJSC
46.209.123.0	46.209.123.255	Respina Networks & Beyond PJSC
46.209.124.0	46.209.125.255	Respina Networks & Beyond PJSC
46.209.126.0	46.209.126.255	Respina Networks & Beyond PJSC
46.209.127.0	46.209.127.255	Respina Networks & Beyond PJSC
46.209.128.0	46.209.128.255	Respina Networks & Beyond PJSC
46.209.129.0	46.209.129.255	Respina Networks & Beyond PJSC
46.209.130.0	46.209.130.255	Respina Networks & Beyond PJSC
46.209.131.0	46.209.131.255	Respina Networks & Beyond PJSC
46.209.132.0	46.209.135.255	Respina Networks & Beyond PJSC
46.209.136.0	46.209.136.255	Respina Networks & Beyond PJSC
46.209.137.0	46.209.137.255	Respina Networks & Beyond PJSC
46.209.138.0	46.209.139.255	Respina Networks & Beyond PJSC
46.209.140.0	46.209.140.255	Respina Networks & Beyond PJSC
46.209.141.0	46.209.141.255	Respina Networks & Beyond PJSC
46.209.142.0	46.209.142.255	Respina Networks & Beyond PJSC
46.209.143.0	46.209.143.255	Respina Networks & Beyond PJSC
46.209.144.0	46.209.144.255	Respina Networks & Beyond PJSC
46.209.145.0	46.209.145.255	Respina Networks & Beyond PJSC
46.209.146.0	46.209.146.255	Respina Networks & Beyond PJSC
46.209.147.0	46.209.147.255	Respina Networks & Beyond PJSC
46.209.148.0	46.209.148.255	Respina Networks & Beyond PJSC
46.209.149.0	46.209.149.255	Respina Networks & Beyond PJSC
46.209.150.0	46.209.151.255	Respina Networks & Beyond PJSC
46.209.152.0	46.209.153.255	Respina Networks & Beyond PJSC
46.209.154.0	46.209.154.255	Respina Networks & Beyond PJSC
46.209.155.0	46.209.155.255	Respina Networks & Beyond PJSC
46.209.156.0	46.209.156.255	Respina Networks & Beyond PJSC
46.209.157.0	46.209.157.255	Respina Networks & Beyond PJSC
46.209.158.0	46.209.158.255	Respina Networks & Beyond PJSC
46.209.159.0	46.209.159.255	Respina Networks & Beyond PJSC
46.209.160.0	46.209.160.255	Respina Networks & Beyond PJSC
46.209.161.0	46.209.161.255	Respina Networks & Beyond PJSC
46.209.162.0	46.209.163.255	Respina Networks & Beyond PJSC
46.209.164.0	46.209.167.255	Respina Networks & Beyond PJSC
46.209.168.0	46.209.175.255	Respina Networks & Beyond PJSC
46.209.176.0	46.209.176.255	Respina Networks & Beyond PJSC
46.209.177.0	46.209.177.255	Respina Networks & Beyond PJSC
46.209.178.0	46.209.179.255	Respina Networks & Beyond PJSC
46.209.180.0	46.209.180.255	Respina Networks & Beyond PJSC
46.209.181.0	46.209.181.255	Respina Networks & Beyond PJSC
46.209.182.0	46.209.183.255	Respina Networks & Beyond PJSC
46.209.184.0	46.209.187.255	Respina Networks & Beyond PJSC
46.209.188.0	46.209.189.255	Respina Networks & Beyond PJSC
46.209.190.0	46.209.190.255	Respina Networks & Beyond PJSC
46.209.191.0	46.209.191.255	Respina Networks & Beyond PJSC
46.209.192.0	46.209.193.255	Respina Networks & Beyond PJSC
46.209.194.0	46.209.194.255	Respina Networks & Beyond PJSC
46.209.195.0	46.209.195.255	Respina Networks & Beyond PJSC
46.209.196.0	46.209.197.255	Respina Networks & Beyond PJSC
46.209.198.0	46.209.198.255	Respina Networks & Beyond PJSC
46.209.199.0	46.209.199.255	Respina Networks & Beyond PJSC
46.209.200.0	46.209.200.255	Respina Networks & Beyond PJSC
46.209.201.0	46.209.201.255	Respina Networks & Beyond PJSC
46.209.202.0	46.209.202.255	Respina Networks & Beyond PJSC
46.209.203.0	46.209.203.255	Respina Networks & Beyond PJSC
46.209.204.0	46.209.204.255	Respina Networks & Beyond PJSC
46.209.205.0	46.209.205.255	Respina Networks & Beyond PJSC
46.209.206.0	46.209.206.255	Respina Networks & Beyond PJSC
46.209.207.0	46.209.207.255	Respina Networks & Beyond PJSC
46.209.208.0	46.209.208.255	Respina Networks & Beyond PJSC
46.209.209.0	46.209.209.255	Respina Networks & Beyond PJSC
46.209.210.0	46.209.210.255	Respina Networks & Beyond PJSC
46.209.211.0	46.209.211.255	Respina Networks & Beyond PJSC
46.209.212.0	46.209.212.255	Respina Networks & Beyond PJSC
46.209.213.0	46.209.213.255	Respina Networks & Beyond PJSC
46.209.214.0	46.209.214.255	Respina Networks & Beyond PJSC
46.209.215.0	46.209.215.255	Respina Networks & Beyond PJSC
46.209.216.0	46.209.216.255	Respina Networks & Beyond PJSC
46.209.217.0	46.209.217.255	Respina Networks & Beyond PJSC
46.209.218.0	46.209.218.255	Respina Networks & Beyond PJSC
46.209.219.0	46.209.219.255	Respina Networks & Beyond PJSC
46.209.220.0	46.209.220.255	Respina Networks & Beyond PJSC
46.209.221.0	46.209.221.255	Respina Networks & Beyond PJSC
46.209.222.0	46.209.222.255	Respina Networks & Beyond PJSC
46.209.223.0	46.209.223.255	Respina Networks & Beyond PJSC
46.209.224.0	46.209.227.255	Respina Networks & Beyond PJSC
46.209.228.0	46.209.229.255	Respina Networks & Beyond PJSC
46.209.230.0	46.209.230.255	Respina Networks & Beyond PJSC
46.209.231.0	46.209.231.255	Respina Networks & Beyond PJSC
46.209.232.0	46.209.232.255	Respina Networks & Beyond PJSC
46.209.233.0	46.209.233.255	Respina Networks & Beyond PJSC
46.209.234.0	46.209.235.255	Respina Networks & Beyond PJSC
46.209.236.0	46.209.236.255	Respina Networks & Beyond PJSC
46.209.237.0	46.209.237.255	Respina Networks & Beyond PJSC
46.209.238.0	46.209.238.255	Respina Networks & Beyond PJSC
46.209.239.0	46.209.239.255	Respina Networks & Beyond PJSC
46.209.240.0	46.209.240.255	Respina Networks & Beyond PJSC
46.209.241.0	46.209.241.255	Respina Networks & Beyond PJSC
46.209.242.0	46.209.243.255	Respina Networks & Beyond PJSC
46.209.244.0	46.209.244.255	Respina Networks & Beyond PJSC
46.209.245.0	46.209.245.255	Respina Networks & Beyond PJSC
46.209.246.0	46.209.246.255	Respina Networks & Beyond PJSC
46.209.247.0	46.209.247.255	Respina Networks & Beyond PJSC
46.209.248.0	46.209.249.255	Respina Networks & Beyond PJSC
46.209.250.0	46.209.250.255	Respina Networks & Beyond PJSC
46.209.251.0	46.209.251.255	Respina Networks & Beyond PJSC
46.209.252.0	46.209.252.255	Respina Networks & Beyond PJSC
46.209.253.0	46.209.253.255	Respina Networks & Beyond PJSC
46.209.254.0	46.209.254.255	Respina Networks & Beyond PJSC
46.209.255.0	46.209.255.255	Respina Networks & Beyond PJSC
46.235.76.0	46.235.77.255	Iran Telecommunication Company Pjs
46.245.0.0	46.245.7.255	Hamara System Tabriz Engineering Company
46.245.8.0	46.245.8.255	Hamara System Tabriz Engineering Company
46.245.9.0	46.245.9.255	Hamara System Tabriz Engineering Company
46.245.10.0	46.245.11.255	Hamara System Tabriz Engineering Company
46.245.12.0	46.245.15.255	Hamara System Tabriz Engineering Company
46.245.16.0	46.245.31.255	Hamara System Tabriz Engineering Company
46.245.32.0	46.245.47.255	Hamara System Tabriz Engineering Company
46.245.48.0	46.245.55.255	Hezardastan Unit Cloud Computing PJSC
46.245.56.0	46.245.63.255	Hamara System Tabriz Engineering Company
46.245.64.0	46.245.67.255	Asiatech Data Transmission Company
46.245.68.0	46.245.68.255	Asiatech Data Transmission Company
46.245.69.0	46.245.69.255	Asiatech Data Transmission Company
46.245.70.0	46.245.70.255	Asiatech Data Transmission Company
46.245.71.0	46.245.71.255	Asiatech Data Transmission Company
46.245.72.0	46.245.75.255	Asiatech Data Transmission Company
46.245.76.0	46.245.76.255	Asiatech Data Transmission Company
46.245.77.0	46.245.77.255	Asiatech Data Transmission Company
46.245.78.0	46.245.79.255	Asiatech Data Transmission Company
46.245.80.0	46.245.81.255	Asiatech Data Transmission Company
46.245.82.0	46.245.82.255	Asiatech Data Transmission Company
46.245.83.0	46.245.83.255	Asiatech Data Transmission Company
46.245.84.0	46.245.84.255	Asiatech Data Transmission Company
46.245.85.0	46.245.85.255	Asiatech Data Transmission Company
46.245.86.0	46.245.86.255	Asiatech Data Transmission Company
46.245.87.0	46.245.87.255	Asiatech Data Transmission Company
46.245.88.0	46.245.91.255	Asiatech Data Transmission Company
46.245.92.0	46.245.93.255	Asiatech Data Transmission Company
46.245.94.0	46.245.94.255	Asiatech Data Transmission Company
46.245.95.0	46.245.95.255	Asiatech Data Transmission Company
46.245.96.0	46.245.96.255	Asiatech Data Transmission Company
46.245.97.0	46.245.97.255	Asiatech Data Transmission Company
46.245.98.0	46.245.99.255	Asiatech Data Transmission Company
46.245.100.0	46.245.103.255	Asiatech Data Transmission Company
46.245.104.0	46.245.107.255	Asiatech Data Transmission Company
46.245.108.0	46.245.111.255	Asiatech Data Transmission Company
46.245.112.0	46.245.113.255	Asiatech Data Transmission Company
46.245.114.0	46.245.114.255	Asiatech Data Transmission Company
46.245.115.0	46.245.115.255	Asiatech Data Transmission Company
46.245.116.0	46.245.116.255	Asiatech Data Transmission Company
46.245.117.0	46.245.117.255	Asiatech Data Transmission Company
46.245.118.0	46.245.118.255	Asiatech Data Transmission Company
46.245.119.0	46.245.119.255	Asiatech Data Transmission Company
46.245.120.0	46.245.123.255	Asiatech Data Transmission Company
46.245.124.0	46.245.124.255	Asiatech Data Transmission Company
46.245.125.0	46.245.125.255	Asiatech Data Transmission Company
46.245.126.0	46.245.127.255	Asiatech Data Transmission Company
46.248.32.0	46.248.63.255	Iran Telecommunication Company Pjs
46.249.96.0	46.249.96.255	Tose Masiryabi Shabake Aria PJSC
46.249.120.0	46.249.123.255	Didehban Net Company PJS
46.249.124.0	46.249.127.255	Didehban Net Company PJS
46.251.224.0	46.251.224.255	Parsan Lin Co. PJS
46.251.226.0	46.251.226.255	Parsan Lin Co. PJS
46.251.237.0	46.251.237.255	Parsan Lin Co. PJS
46.255.216.0	46.255.216.255	United Nations High Commissioner for Refugees
46.255.217.0	46.255.217.255	United Nations High Commissioner for Refugees
46.255.218.0	46.255.219.255	United Nations High Commissioner for Refugees
46.255.220.0	46.255.223.255	United Nations High Commissioner for Refugees
62.3.14.0	62.3.14.255	Patron Technology Persia Ltd
62.3.41.0	62.3.41.255	Vandad Vira Hooman LLC
62.3.42.0	62.3.42.255	Green Web Samaneh Novin PJSC
62.60.128.0	62.60.128.255	Noavaran Shabakeh Sabz Mehregan (Ltd.)
62.60.129.0	62.60.129.255	Noavaran Shabakeh Sabz Mehregan (Ltd.)
62.60.132.0	62.60.132.255	Iranian Research Organization for Science & Technology
62.60.133.0	62.60.133.255	Noavaran Shabakeh Sabz Mehregan (Ltd.)
62.60.135.0	62.60.135.255	Iranian Research Organization for Science & Technology
62.60.136.0	62.60.136.255	Iranian Research Organization for Science & Technology
62.60.137.0	62.60.137.255	Iranian Research Organization for Science & Technology
62.60.139.0	62.60.139.255	Iranian Research Organization for Science & Technology
62.60.140.0	62.60.140.255	Iranian Research Organization for Science & Technology
62.60.141.0	62.60.141.255	Abr Baran Information Technologists LLC
62.60.142.0	62.60.142.255	Iranian Research Organization for Science & Technology
62.60.143.0	62.60.143.255	Parsway Shomal Company Ltd
62.60.145.0	62.60.145.255	Iranian Research Organization for Science & Technology
62.60.146.0	62.60.146.255	Abr Baran Information Technologists LLC
62.60.154.0	62.60.154.255	Iranian Research Organization for Science & Technology
62.60.162.0	62.60.162.255	Asre Pardazeshe Ettelaate Amin Institute
62.60.163.0	62.60.163.255	Iranian Research Organization for Science & Technology
62.60.164.0	62.60.164.255	Noavaran Shabakeh Sabz Mehregan (Ltd.)
62.60.165.0	62.60.165.255	Noavaran Shabakeh Sabz Mehregan (Ltd.)
62.60.166.0	62.60.166.255	Noavaran Shabakeh Sabz Mehregan (Ltd.)
62.60.167.0	62.60.167.255	Noavaran Shabakeh Sabz Mehregan (Ltd.)
62.60.168.0	62.60.175.255	Iranian Research Organization for Science & Technology
62.60.176.0	62.60.183.255	Iranian Research Organization for Science & Technology
62.60.184.0	62.60.185.255	Iranian Research Organization for Science & Technology
62.60.188.0	62.60.189.255	Iranian Research Organization for Science & Technology
62.60.191.0	62.60.191.255	Iranian Research Organization for Science & Technology
62.60.192.0	62.60.193.255	Iranian Research Organization for Science & Technology
62.60.194.0	62.60.194.255	Tose'e Shabakeh Azad Ltd
62.60.195.0	62.60.195.255	Iranian Research Organization for Science & Technology
62.60.196.0	62.60.196.255	Asre Pardazeshe Ettelaate Amin Institute
62.60.197.0	62.60.197.255	Asre Pardazeshe Ettelaate Amin Institute
62.60.198.0	62.60.198.255	Tose'e Shabakeh Azad Ltd
62.60.199.0	62.60.199.255	Asre Pardazeshe Ettelaate Amin Institute
62.60.200.0	62.60.200.255	Noavaran Shabakeh Sabz Mehregan (Ltd.)
62.60.201.0	62.60.201.255	Noavaran Shabakeh Sabz Mehregan (Ltd.)
62.60.202.0	62.60.202.255	Noavaran Shabakeh Sabz Mehregan (Ltd.)
62.60.203.0	62.60.203.255	Noavaran Shabakeh Sabz Mehregan (Ltd.)
62.60.204.0	62.60.204.255	Noavaran Shabakeh Sabz Mehregan (Ltd.)
62.60.205.0	62.60.205.255	Noavaran Shabakeh Sabz Mehregan (Ltd.)
62.60.206.0	62.60.206.255	Noavaran Shabakeh Sabz Mehregan (Ltd.)
62.60.207.0	62.60.207.255	Noavaran Shabakeh Sabz Mehregan (Ltd.)
62.60.208.0	62.60.208.255	Iranian Research Organization for Science & Technology
62.60.209.0	62.60.209.255	Noavaran Shabakeh Sabz Mehregan (Ltd.)
62.60.210.0	62.60.210.255	Noavaran Shabakeh Sabz Mehregan (Ltd.)
62.60.211.0	62.60.211.255	Noavaran Shabakeh Sabz Mehregan (Ltd.)
62.60.212.0	62.60.212.255	Noavaran Shabakeh Sabz Mehregan (Ltd.)
62.60.213.0	62.60.213.255	Noavaran Shabakeh Sabz Mehregan (Ltd.)
62.60.214.0	62.60.214.255	Noavaran Shabakeh Sabz Mehregan (Ltd.)
62.60.215.0	62.60.215.255	Noavaran Shabakeh Sabz Mehregan (Ltd.)
62.60.220.0	62.60.221.255	Iranian Research Organization for Science & Technology
62.60.222.0	62.60.222.255	FanAvaran Mihan Mizban PJSC
62.60.223.0	62.60.223.255	FanAvaran Mihan Mizban PJSC
62.60.224.0	62.60.225.255	Iranian Research Organization for Science & Technology
62.60.240.0	62.60.243.255	Iranian Research Organization for Science & Technology
62.102.128.0	62.102.143.255	Mobile Communication Company of Iran Plc
62.106.95.0	62.106.95.255	Vandad Vira Hooman LLC
62.193.0.0	62.193.1.255	DP Iran PLC
62.193.2.0	62.193.2.255	DP Iran PLC
62.193.3.0	62.193.3.255	DP Iran PLC
62.193.4.0	62.193.4.255	DP Iran PLC
62.193.5.0	62.193.5.255	DP Iran PLC
62.193.6.0	62.193.6.255	DP Iran PLC
62.193.7.0	62.193.7.255	DP Iran PLC
62.193.8.0	62.193.8.255	DP Iran PLC
62.193.9.0	62.193.9.255	DP Iran PLC
62.193.10.0	62.193.11.255	DP Iran PLC
62.193.12.0	62.193.12.255	DP Iran PLC
62.193.13.0	62.193.13.255	DP Iran PLC
62.193.14.0	62.193.15.255	DP Iran PLC
62.193.16.0	62.193.16.255	DP Iran PLC
62.193.17.0	62.193.17.255	DP Iran PLC
62.193.18.0	62.193.18.255	DP Iran PLC
62.193.19.0	62.193.19.255	DP Iran PLC
62.193.20.0	62.193.20.255	DP Iran PLC
62.193.21.0	62.193.21.255	DP Iran PLC
62.193.22.0	62.193.23.255	DP Iran PLC
62.193.24.0	62.193.31.255	DP Iran PLC
62.204.61.0	62.204.61.255	Saba Abr Mizban LLC
62.220.96.0	62.220.99.255	Soroush Rasanheh Company Ltd
62.220.100.0	62.220.101.255	Soroush Rasanheh Company Ltd
62.220.102.0	62.220.102.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
62.220.103.0	62.220.103.255	Soroush Rasanheh Company Ltd
62.220.104.0	62.220.107.255	Soroush Rasanheh Company Ltd
62.220.108.0	62.220.108.255	Iran Chamber of Commerce Industries Mines and Agriculture
62.220.109.0	62.220.109.255	Soroush Rasanheh Company Ltd
62.220.110.0	62.220.111.255	Soroush Rasanheh Company Ltd
62.220.112.0	62.220.112.255	Respina Networks & Beyond PJSC
62.220.113.0	62.220.113.255	Respina Networks & Beyond PJSC
62.220.114.0	62.220.115.255	Respina Networks & Beyond PJSC
62.220.116.0	62.220.116.255	Respina Networks & Beyond PJSC
62.220.117.0	62.220.117.255	Respina Networks & Beyond PJSC
62.220.118.0	62.220.119.255	Respina Networks & Beyond PJSC
62.220.120.0	62.220.120.255	Giti Secure Cloud LLC
62.220.121.0	62.220.121.255	Soroush Rasanheh Company Ltd
62.220.122.0	62.220.122.255	Dade Pardazi Mobinhost Co Ltd
62.220.123.0	62.220.123.255	Dade Pardazi Mobinhost Co Ltd
62.220.124.0	62.220.124.255	Dade Pardazi Mobinhost Co Ltd
62.220.125.0	62.220.125.255	Dade Pardazi Mobinhost Co Ltd
62.220.126.0	62.220.126.255	Dade Pardazi Mobinhost Co Ltd
62.220.127.0	62.220.127.255	Dade Pardazi Mobinhost Co Ltd
66.79.96.0	66.79.103.255	Iran Telecommunication Company Pjs
66.79.104.0	66.79.104.255	Iran Telecommunication Company Pjs
66.79.105.0	66.79.105.255	Iran Telecommunication Company Pjs
66.79.106.0	66.79.107.255	Iran Telecommunication Company Pjs
66.79.108.0	66.79.111.255	Iran Telecommunication Company Pjs
66.79.112.0	66.79.113.255	Iran Telecommunication Company Pjs
66.79.114.0	66.79.115.255	Iran Telecommunication Company Pjs
66.79.116.0	66.79.117.255	Iran Telecommunication Company Pjs
66.79.118.0	66.79.119.255	Iran Telecommunication Company Pjs
66.79.120.0	66.79.127.255	Iran Telecommunication Company Pjs
69.194.64.0	69.194.79.255	Mobile Communication Company of Iran Plc
69.194.80.0	69.194.95.255	Mobile Communication Company of Iran Plc
69.194.96.0	69.194.111.255	Mobile Communication Company of Iran Plc
69.194.112.0	69.194.127.255	Mobile Communication Company of Iran Plc
77.36.128.0	77.36.143.255	IRIB (Islamic Republic of Iran Broadcasting)
77.36.144.0	77.36.159.255	IRIB (Islamic Republic of Iran Broadcasting)
77.36.160.0	77.36.175.255	IRIB (Islamic Republic of Iran Broadcasting)
77.36.176.0	77.36.191.255	IRIB (Islamic Republic of Iran Broadcasting)
77.36.192.0	77.36.207.255	IRIB (Islamic Republic of Iran Broadcasting)
77.36.208.0	77.36.223.255	IRIB (Islamic Republic of Iran Broadcasting)
77.36.224.0	77.36.239.255	IRIB (Islamic Republic of Iran Broadcasting)
77.36.240.0	77.36.255.255	IRIB (Islamic Republic of Iran Broadcasting)
77.72.80.0	77.72.80.255	Saba Abr Mizban LLC
77.74.202.0	77.74.202.255	Toesegaran Shabakeh Arseh Novin Ltd
77.77.64.0	77.77.71.255	Parsan Lin Co. PJS
77.77.72.0	77.77.75.255	Parsan Lin Co. PJS
77.77.76.0	77.77.76.255	Parsan Lin Co. PJS
77.77.77.0	77.77.77.255	Parsan Lin Co. PJS
77.77.78.0	77.77.79.255	Parsan Lin Co. PJS
77.77.80.0	77.77.95.255	Parsan Lin Co. PJS
77.77.96.0	77.77.127.255	Parsan Lin Co. PJS
77.81.32.0	77.81.47.255	Iran Telecommunication Company Pjs
77.81.128.0	77.81.135.255	Iran Telecommunication Company Pjs
77.81.144.0	77.81.159.255	Iran Telecommunication Company Pjs
77.81.192.0	77.81.207.255	Rightel Communication Service Company PJS
77.81.208.0	77.81.223.255	Rightel Communication Service Company PJS
77.95.219.0	77.95.219.255	Cooperative Afra ertebatat-e-sabet-e Rasa Co
77.104.64.0	77.104.64.255	Respina Networks & Beyond PJSC
77.104.65.0	77.104.65.255	Respina Networks & Beyond PJSC
77.104.66.0	77.104.66.255	Respina Networks & Beyond PJSC
77.104.67.0	77.104.67.255	Respina Networks & Beyond PJSC
77.104.68.0	77.104.68.255	Respina Networks & Beyond PJSC
77.104.69.0	77.104.69.255	Respina Networks & Beyond PJSC
77.104.70.0	77.104.70.255	Respina Networks & Beyond PJSC
77.104.71.0	77.104.71.255	Respina Networks & Beyond PJSC
77.104.72.0	77.104.72.255	Respina Networks & Beyond PJSC
77.104.73.0	77.104.73.255	Respina Networks & Beyond PJSC
77.104.74.0	77.104.74.255	Respina Networks & Beyond PJSC
77.104.75.0	77.104.75.255	Respina Networks & Beyond PJSC
77.104.76.0	77.104.76.255	Respina Networks & Beyond PJSC
77.104.77.0	77.104.77.255	Respina Networks & Beyond PJSC
77.104.78.0	77.104.78.255	Respina Networks & Beyond PJSC
77.104.79.0	77.104.79.255	Respina Networks & Beyond PJSC
77.104.80.0	77.104.80.255	Respina Networks & Beyond PJSC
77.104.81.0	77.104.81.255	Respina Networks & Beyond PJSC
77.104.82.0	77.104.82.255	Respina Networks & Beyond PJSC
77.104.83.0	77.104.83.255	Respina Networks & Beyond PJSC
77.104.84.0	77.104.84.255	Respina Networks & Beyond PJSC
77.104.85.0	77.104.85.255	Respina Networks & Beyond PJSC
77.104.86.0	77.104.86.255	Respina Networks & Beyond PJSC
77.104.87.0	77.104.87.255	Respina Networks & Beyond PJSC
77.104.88.0	77.104.88.255	Respina Networks & Beyond PJSC
77.104.89.0	77.104.89.255	Respina Networks & Beyond PJSC
77.104.90.0	77.104.90.255	Respina Networks & Beyond PJSC
77.104.91.0	77.104.91.255	Respina Networks & Beyond PJSC
77.104.92.0	77.104.92.255	Padiz Dadeh Resan Pjsc
77.104.93.0	77.104.93.255	Respina Networks & Beyond PJSC
77.104.94.0	77.104.94.255	Toesegaran Shabakeh Arseh Novin Ltd
77.104.95.0	77.104.95.255	Respina Networks & Beyond PJSC
77.104.96.0	77.104.96.255	Respina Networks & Beyond PJSC
77.104.97.0	77.104.97.255	Respina Networks & Beyond PJSC
77.104.98.0	77.104.98.255	Respina Networks & Beyond PJSC
77.104.99.0	77.104.99.255	Respina Networks & Beyond PJSC
77.104.100.0	77.104.100.255	Respina Networks & Beyond PJSC
77.104.101.0	77.104.101.255	Respina Networks & Beyond PJSC
77.104.102.0	77.104.102.255	Respina Networks & Beyond PJSC
77.104.103.0	77.104.103.255	Respina Networks & Beyond PJSC
77.104.104.0	77.104.104.255	Respina Networks & Beyond PJSC
77.104.105.0	77.104.105.255	Respina Networks & Beyond PJSC
77.104.106.0	77.104.106.255	Respina Networks & Beyond PJSC
77.104.107.0	77.104.107.255	Respina Networks & Beyond PJSC
77.104.108.0	77.104.109.255	Respina Networks & Beyond PJSC
77.104.110.0	77.104.110.255	Respina Networks & Beyond PJSC
77.104.111.0	77.104.111.255	Respina Networks & Beyond PJSC
77.104.112.0	77.104.112.255	Respina Networks & Beyond PJSC
77.104.113.0	77.104.113.255	Respina Networks & Beyond PJSC
77.104.114.0	77.104.114.255	Respina Networks & Beyond PJSC
77.104.115.0	77.104.115.255	Respina Networks & Beyond PJSC
77.104.116.0	77.104.116.255	Respina Networks & Beyond PJSC
77.104.117.0	77.104.117.255	Respina Networks & Beyond PJSC
77.104.118.0	77.104.118.255	Respina Networks & Beyond PJSC
77.104.119.0	77.104.119.255	Respina Networks & Beyond PJSC
77.104.120.0	77.104.123.255	Respina Networks & Beyond PJSC
77.104.124.0	77.104.124.255	Pars Tamin Navid Company (Private Joint-Stock)
77.104.125.0	77.104.125.255	Respina Networks & Beyond PJSC
77.104.126.0	77.104.126.255	Respina Networks & Beyond PJSC
77.104.127.0	77.104.127.255	Respina Networks & Beyond PJSC
77.237.64.0	77.237.64.255	Avaye Etminan Asset Management PJSC
77.237.65.0	77.237.65.255	Respina Networks & Beyond PJSC
77.237.66.0	77.237.66.255	Rasaneh Avabarid Private Joint Stock Company
77.237.67.0	77.237.67.255	Respina Networks & Beyond PJSC
77.237.68.0	77.237.68.255	Respina Networks & Beyond PJSC
77.237.69.0	77.237.69.255	Respina Networks & Beyond PJSC
77.237.70.0	77.237.70.255	Respina Networks & Beyond PJSC
77.237.71.0	77.237.71.255	Respina Networks & Beyond PJSC
77.237.72.0	77.237.72.255	Respina Networks & Beyond PJSC
77.237.73.0	77.237.73.255	Toesegaran Shabakeh Arseh Novin Ltd
77.237.74.0	77.237.74.255	Respina Networks & Beyond PJSC
77.237.75.0	77.237.75.255	Respina Networks & Beyond PJSC
77.237.76.0	77.237.76.255	Respina Networks & Beyond PJSC
77.237.77.0	77.237.77.255	Respina Networks & Beyond PJSC
77.237.78.0	77.237.78.255	Respina Networks & Beyond PJSC
77.237.79.0	77.237.79.255	Iran Tax Affairs Organisation
77.237.80.0	77.237.80.255	Gostaresh Energy Pasargad Company (PJS)
77.237.81.0	77.237.81.255	Respina Networks & Beyond PJSC
77.237.82.0	77.237.82.255	Respina Networks & Beyond PJSC
77.237.83.0	77.237.83.255	Respina Networks & Beyond PJSC
77.237.84.0	77.237.84.255	Respina Networks & Beyond PJSC
77.237.85.0	77.237.85.255	Respina Networks & Beyond PJSC
77.237.86.0	77.237.86.255	Respina Networks & Beyond PJSC
77.237.87.0	77.237.87.255	Respina Networks & Beyond PJSC
77.237.88.0	77.237.88.255	Respina Networks & Beyond PJSC
77.237.89.0	77.237.89.255	Respina Networks & Beyond PJSC
77.237.90.0	77.237.90.255	Respina Networks & Beyond PJSC
77.237.91.0	77.237.91.255	Respina Networks & Beyond PJSC
77.237.92.0	77.237.93.255	Respina Networks & Beyond PJSC
77.237.94.0	77.237.94.255	Respina Networks & Beyond PJSC
77.237.95.0	77.237.95.255	Respina Networks & Beyond PJSC
77.237.160.0	77.237.191.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
77.238.104.0	77.238.107.255	Asiatech Data Transmission Company
77.238.108.0	77.238.108.255	Asiatech Data Transmission Company
77.238.109.0	77.238.109.255	Asiatech Data Transmission Company
77.238.110.0	77.238.110.255	Asiatech Data Transmission Company
77.238.111.0	77.238.111.255	Asiatech Data Transmission Company
77.238.112.0	77.238.119.255	Iran Telecommunication Company Pjs
77.238.120.0	77.238.120.255	Asiatech Data Transmission Company
77.238.121.0	77.238.121.255	Asiatech Data Transmission Company
77.238.122.0	77.238.122.255	Asiatech Data Transmission Company
77.238.123.0	77.238.123.255	Asiatech Data Transmission Company
77.238.124.0	77.238.127.255	Iran Telecommunication Company Pjs
77.245.224.0	77.245.224.255	Research Institute of Petroleum Industry
77.245.225.0	77.245.225.255	Research Institute of Petroleum Industry
77.245.226.0	77.245.226.255	Research Institute of Petroleum Industry
77.245.227.0	77.245.227.255	Research Institute of Petroleum Industry
77.245.228.0	77.245.228.255	Research Institute of Petroleum Industry
77.245.229.0	77.245.229.255	Research Institute of Petroleum Industry
77.245.230.0	77.245.230.255	Research Institute of Petroleum Industry
77.245.231.0	77.245.231.255	Research Institute of Petroleum Industry
77.245.232.0	77.245.232.255	Research Institute of Petroleum Industry
77.245.233.0	77.245.233.255	Research Institute of Petroleum Industry
77.245.234.0	77.245.234.255	Research Institute of Petroleum Industry
77.245.235.0	77.245.235.255	Research Institute of Petroleum Industry
77.245.236.0	77.245.236.255	Research Institute of Petroleum Industry
77.245.237.0	77.245.237.255	Research Institute of Petroleum Industry
77.245.238.0	77.245.239.255	Research Institute of Petroleum Industry
78.31.232.0	78.31.235.255	Abramad Technological Infrastructures Development Company PJS
78.38.0.0	78.38.7.255	Iran Telecommunication Company Pjs
78.38.8.0	78.38.15.255	Telecommunication Infrastructure Company
78.38.16.0	78.38.23.255	Iran Telecommunication Company Pjs
78.38.24.0	78.38.25.255	Iran Telecommunication Company Pjs
78.38.26.0	78.38.27.255	Iran Telecommunication Company Pjs
78.38.28.0	78.38.31.255	Iran Telecommunication Company Pjs
78.38.32.0	78.38.39.255	Iran Telecommunication Company Pjs
78.38.40.0	78.38.41.255	Iran Telecommunication Company Pjs
78.38.42.0	78.38.43.255	Iran Telecommunication Company Pjs
78.38.44.0	78.38.47.255	Iran Telecommunication Company Pjs
78.38.48.0	78.38.49.255	Iran Telecommunication Company Pjs
78.38.50.0	78.38.50.255	Iran Telecommunication Company Pjs
78.38.51.0	78.38.51.255	Iran Telecommunication Company Pjs
78.38.52.0	78.38.52.255	Iran Telecommunication Company Pjs
78.38.53.0	78.38.53.255	Iran Telecommunication Company Pjs
78.38.54.0	78.38.54.255	Iran Telecommunication Company Pjs
78.38.55.0	78.38.55.255	Iran Telecommunication Company Pjs
78.38.56.0	78.38.63.255	Iran Telecommunication Company Pjs
78.38.64.0	78.38.71.255	Iran Telecommunication Company Pjs
78.38.72.0	78.38.79.255	Iran Telecommunication Company Pjs
78.38.80.0	78.38.87.255	Iran Telecommunication Company Pjs
78.38.88.0	78.38.95.255	Iran Telecommunication Company Pjs
78.38.96.0	78.38.99.255	Iran Telecommunication Company Pjs
78.38.100.0	78.38.103.255	Iran Telecommunication Company Pjs
78.38.104.0	78.38.105.255	Iran Telecommunication Company Pjs
78.38.106.0	78.38.107.255	Iran Telecommunication Company Pjs
78.38.108.0	78.38.111.255	Iran Telecommunication Company Pjs
78.38.112.0	78.38.119.255	Iran Telecommunication Company Pjs
78.38.120.0	78.38.127.255	Iran Telecommunication Company Pjs
78.38.128.0	78.38.143.255	Iran Telecommunication Company Pjs
78.38.144.0	78.38.147.255	Iran Telecommunication Company Pjs
78.38.148.0	78.38.151.255	Iran Telecommunication Company Pjs
78.38.152.0	78.38.159.255	Iran Telecommunication Company Pjs
78.38.160.0	78.38.167.255	Iran Telecommunication Company Pjs
78.38.168.0	78.38.175.255	Iran Telecommunication Company Pjs
78.38.176.0	78.38.191.255	Iran Telecommunication Company Pjs
78.38.192.0	78.38.207.255	Iran Telecommunication Company Pjs
78.38.208.0	78.38.215.255	Iran Telecommunication Company Pjs
78.38.216.0	78.38.217.255	Iran Telecommunication Company Pjs
78.38.218.0	78.38.218.255	Iran Telecommunication Company Pjs
78.38.219.0	78.38.219.255	Iran Telecommunication Company Pjs
78.38.220.0	78.38.223.255	Iran Telecommunication Company Pjs
78.38.224.0	78.38.225.255	Iran Telecommunication Company Pjs
78.38.226.0	78.38.227.255	Iran Telecommunication Company Pjs
78.38.228.0	78.38.229.255	Iran Telecommunication Company Pjs
78.38.230.0	78.38.231.255	Iran Telecommunication Company Pjs
78.38.232.0	78.38.233.255	Iran Telecommunication Company Pjs
78.38.234.0	78.38.235.255	Iran Telecommunication Company Pjs
78.38.236.0	78.38.236.255	Iran Telecommunication Company Pjs
78.38.237.0	78.38.237.255	Iran Information Technology Company PJSC
78.38.238.0	78.38.238.255	Telecommunication Infrastructure Company
78.38.239.0	78.38.239.255	Iran Information Technology Company PJSC
78.38.240.0	78.38.240.255	Iran Information Technology Company PJSC
78.38.241.0	78.38.241.255	Iran Information Technology Company PJSC
78.38.242.0	78.38.242.255	Iran Information Technology Company PJSC
78.38.243.0	78.38.243.255	Respina Networks & Beyond PJSC
78.38.244.0	78.38.244.255	Iran Information Technology Company PJSC
78.38.245.0	78.38.245.255	Iran Information Technology Company PJSC
78.38.246.0	78.38.246.255	Respina Networks & Beyond PJSC
78.38.247.0	78.38.247.255	Iran Information Technology Company PJSC
78.38.248.0	78.38.248.255	Respina Networks & Beyond PJSC
78.38.249.0	78.38.249.255	Telecommunication Infrastructure Company
78.38.250.0	78.38.250.255	Respina Networks & Beyond PJSC
78.38.251.0	78.38.251.255	Respina Networks & Beyond PJSC
78.38.252.0	78.38.253.255	Telecommunication Infrastructure Company
78.38.254.0	78.38.254.255	Respina Networks & Beyond PJSC
78.38.255.0	78.38.255.255	Iran Information Technology Company PJSC
78.39.0.0	78.39.15.255	Iran Telecommunication Company Pjs
78.39.16.0	78.39.23.255	Iran Telecommunication Company Pjs
78.39.24.0	78.39.31.255	Iran Telecommunication Company Pjs
78.39.32.0	78.39.39.255	Iran Telecommunication Company Pjs
78.39.40.0	78.39.40.255	Respina Networks & Beyond PJSC
78.39.41.0	78.39.41.255	Iran Telecommunication Company Pjs
78.39.42.0	78.39.42.255	Iran Telecommunication Company Pjs
78.39.43.0	78.39.43.255	Respina Networks & Beyond PJSC
78.39.44.0	78.39.45.255	Iran Telecommunication Company Pjs
78.39.46.0	78.39.46.255	Telecommunication Infrastructure Company
78.39.47.0	78.39.47.255	Respina Networks & Beyond PJSC
78.39.48.0	78.39.48.255	Respina Networks & Beyond PJSC
78.39.49.0	78.39.49.255	Respina Networks & Beyond PJSC
78.39.50.0	78.39.50.255	Respina Networks & Beyond PJSC
78.39.51.0	78.39.51.255	Telecommunication Infrastructure Company
78.39.52.0	78.39.52.255	Iran Telecommunication Company Pjs
78.39.53.0	78.39.53.255	Respina Networks & Beyond PJSC
78.39.54.0	78.39.54.255	Respina Networks & Beyond PJSC
78.39.55.0	78.39.55.255	Respina Networks & Beyond PJSC
78.39.56.0	78.39.56.255	Respina Networks & Beyond PJSC
78.39.57.0	78.39.57.255	Respina Networks & Beyond PJSC
78.39.58.0	78.39.58.255	Respina Networks & Beyond PJSC
78.39.59.0	78.39.59.255	Respina Networks & Beyond PJSC
78.39.60.0	78.39.61.255	Iran Telecommunication Company Pjs
78.39.62.0	78.39.62.255	Respina Networks & Beyond PJSC
78.39.63.0	78.39.63.255	Telecommunication Infrastructure Company
78.39.64.0	78.39.79.255	Iran Telecommunication Company Pjs
78.39.80.0	78.39.87.255	Iran Telecommunication Company Pjs
78.39.88.0	78.39.95.255	Iran Telecommunication Company Pjs
78.39.96.0	78.39.103.255	Iran Telecommunication Company Pjs
78.39.104.0	78.39.111.255	Iran Telecommunication Company Pjs
78.39.112.0	78.39.115.255	Iran Telecommunication Company Pjs
78.39.116.0	78.39.117.255	Iran Telecommunication Company Pjs
78.39.118.0	78.39.119.255	Iran Telecommunication Company Pjs
78.39.120.0	78.39.127.255	Iran Telecommunication Company Pjs
78.39.128.0	78.39.128.255	Iran Information Technology Company PJSC
78.39.129.0	78.39.129.255	Iran Telecommunication Company Pjs
78.39.130.0	78.39.131.255	Iran Telecommunication Company Pjs
78.39.132.0	78.39.135.255	Iran Telecommunication Company Pjs
78.39.136.0	78.39.136.255	Iran Telecommunication Company Pjs
78.39.137.0	78.39.137.255	Iran Telecommunication Company Pjs
78.39.138.0	78.39.139.255	Iran Telecommunication Company Pjs
78.39.140.0	78.39.141.255	Iran Telecommunication Company Pjs
78.39.142.0	78.39.143.255	Iran Telecommunication Company Pjs
78.39.144.0	78.39.147.255	Iran Telecommunication Company Pjs
78.39.148.0	78.39.151.255	Iran Telecommunication Company Pjs
78.39.152.0	78.39.152.255	Iran Information Technology Company PJSC
78.39.153.0	78.39.153.255	Iran Information Technology Company PJSC
78.39.154.0	78.39.154.255	Iran Telecommunication Company Pjs
78.39.155.0	78.39.155.255	Telecommunication Infrastructure Company
78.39.156.0	78.39.156.255	Telecommunication Infrastructure Company
78.39.157.0	78.39.157.255	Iran Information Technology Company PJSC
78.39.158.0	78.39.158.255	Iran Telecommunication Company Pjs
78.39.159.0	78.39.159.255	Iran Telecommunication Company Pjs
78.39.160.0	78.39.167.255	Iran Telecommunication Company Pjs
78.39.168.0	78.39.171.255	Iran Telecommunication Company Pjs
78.39.172.0	78.39.173.255	Iran Telecommunication Company Pjs
78.39.174.0	78.39.174.255	Iran Telecommunication Company Pjs
78.39.175.0	78.39.175.255	Iran Telecommunication Company Pjs
78.39.176.0	78.39.177.255	Iran Telecommunication Company Pjs
78.39.178.0	78.39.179.255	Iran Telecommunication Company Pjs
78.39.180.0	78.39.180.255	Iran Telecommunication Company Pjs
78.39.181.0	78.39.181.255	Iran Telecommunication Company Pjs
78.39.182.0	78.39.182.255	Iran Telecommunication Company Pjs
78.39.183.0	78.39.183.255	Iran Telecommunication Company Pjs
78.39.184.0	78.39.184.255	Iran Telecommunication Company Pjs
78.39.185.0	78.39.185.255	Telecommunication Infrastructure Company
78.39.186.0	78.39.187.255	Iran Telecommunication Company Pjs
78.39.188.0	78.39.191.255	Iran Telecommunication Company Pjs
78.39.192.0	78.39.199.255	Iran Telecommunication Company Pjs
78.39.200.0	78.39.203.255	Iran Telecommunication Company Pjs
78.39.204.0	78.39.207.255	Iran Telecommunication Company Pjs
78.39.208.0	78.39.215.255	Iran Telecommunication Company Pjs
78.39.216.0	78.39.219.255	Iran Telecommunication Company Pjs
78.39.220.0	78.39.220.255	Iran Telecommunication Company Pjs
78.39.221.0	78.39.221.255	Iran Telecommunication Company Pjs
78.39.222.0	78.39.223.255	Iran Telecommunication Company Pjs
78.39.224.0	78.39.231.255	Iran Telecommunication Company Pjs
78.39.232.0	78.39.235.255	Iran Telecommunication Company Pjs
78.39.236.0	78.39.239.255	Iran Telecommunication Company Pjs
78.39.240.0	78.39.243.255	Iran Telecommunication Company Pjs
78.39.244.0	78.39.245.255	Iran Telecommunication Company Pjs
78.39.246.0	78.39.246.255	Iran Telecommunication Company Pjs
78.39.247.0	78.39.247.255	Iran Telecommunication Company Pjs
78.39.248.0	78.39.251.255	Iran Telecommunication Company Pjs
78.39.252.0	78.39.253.255	Iran Telecommunication Company Pjs
78.39.254.0	78.39.255.255	Iran Telecommunication Company Pjs
78.41.61.0	78.41.61.255	Tebyan-e-Noor Cultural-Artistic Institute
78.41.62.0	78.41.62.255	Shabakeh Dadehaye Houshmand Iranian LLC
78.41.137.0	78.41.137.255	Foojan Cloud Computing Co. PJS
78.109.192.0	78.109.192.255	Afranet
78.109.193.0	78.109.193.255	Afranet
78.109.194.0	78.109.194.255	Afranet
78.109.195.0	78.109.195.255	Afranet
78.109.196.0	78.109.196.255	Afranet
78.109.197.0	78.109.197.255	Afranet
78.109.198.0	78.109.198.255	Afranet
78.109.199.0	78.109.199.255	Afranet
78.109.200.0	78.109.200.255	Afranet
78.109.201.0	78.109.201.255	Afranet
78.109.202.0	78.109.202.255	Afranet
78.109.203.0	78.109.203.255	Afranet
78.109.204.0	78.109.204.255	Afranet
78.109.205.0	78.109.205.255	Afranet
78.109.206.0	78.109.206.255	Afranet
78.109.207.0	78.109.207.255	Hezardastan Unit Cloud Computing PJSC
78.110.112.0	78.110.119.255	Asre Dadeha Asiatech
78.110.120.0	78.110.120.255	Sefroyek Pardaz Engineering PJSC
78.110.121.0	78.110.121.255	Sefroyek Pardaz Engineering PJSC
78.110.122.0	78.110.122.255	Sefroyek Pardaz Engineering PJSC
78.110.123.0	78.110.123.255	Sefroyek Pardaz Engineering PJSC
78.110.124.0	78.110.127.255	Asiatech Data Transmission Company
78.111.0.0	78.111.0.255	Moasese Gostaresh Etelaat Va Ertebatat Farhangi Neda Rayaneh
78.111.2.0	78.111.2.255	Moasese Gostaresh Etelaat Va Ertebatat Farhangi Neda Rayaneh
78.111.3.0	78.111.3.255	Moasese Gostaresh Etelaat Va Ertebatat Farhangi Neda Rayaneh
78.111.4.0	78.111.4.255	Moasese Gostaresh Etelaat Va Ertebatat Farhangi Neda Rayaneh
78.111.5.0	78.111.5.255	Hezardastan Unit Cloud Computing PJSC
78.111.6.0	78.111.6.255	Hezardastan Unit Cloud Computing PJSC
78.111.7.0	78.111.7.255	Towse'eh Sarmayeh Gozari Entekhab Group PJSC
78.111.8.0	78.111.8.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
78.111.9.0	78.111.9.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
78.111.10.0	78.111.11.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
78.111.12.0	78.111.13.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
78.111.14.0	78.111.14.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
78.111.15.0	78.111.15.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
78.154.32.0	78.154.63.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
78.157.32.0	78.157.32.255	Dade Samane Fanava Company (PJS)
78.157.33.0	78.157.33.255	Fanava Group
78.157.34.0	78.157.34.255	Dade Samane Fanava Company (PJS)
78.157.35.0	78.157.35.255	Dade Samane Fanava Company (PJS)
78.157.36.0	78.157.36.255	Dade Samane Fanava Company (PJS)
78.157.37.0	78.157.37.255	Dade Samane Fanava Company (PJS)
78.157.38.0	78.157.38.255	Dade Samane Fanava Company (PJS)
78.157.39.0	78.157.39.255	Dade Samane Fanava Company (PJS)
78.157.40.0	78.157.40.255	Dade Samane Fanava Company (PJS)
78.157.41.0	78.157.41.255	Dade Samane Fanava Company (PJS)
78.157.42.0	78.157.42.255	Dade Samane Fanava Company (PJS)
78.157.43.0	78.157.43.255	Dade Samane Fanava Company (PJS)
78.157.44.0	78.157.44.255	Dade Samane Fanava Company (PJS)
78.157.45.0	78.157.45.255	Dade Samane Fanava Company (PJS)
78.157.46.0	78.157.46.255	Dade Samane Fanava Company (PJS)
78.157.47.0	78.157.47.255	Dade Samane Fanava Company (PJS)
78.157.48.0	78.157.48.255	Dade Samane Fanava Company (PJS)
78.157.49.0	78.157.49.255	Dade Samane Fanava Company (PJS)
78.157.50.0	78.157.50.255	Dade Samane Fanava Company (PJS)
78.157.51.0	78.157.51.255	Dade Samane Fanava Company (PJS)
78.157.52.0	78.157.52.255	Dade Samane Fanava Company (PJS)
78.157.53.0	78.157.53.255	Dade Samane Fanava Company (PJS)
78.157.54.0	78.157.55.255	Dade Samane Fanava Company (PJS)
78.157.56.0	78.157.56.255	Dade Samane Fanava Company (PJS)
78.157.57.0	78.157.57.255	Dade Samane Fanava Company (PJS)
78.157.58.0	78.157.58.255	Dade Samane Fanava Company (PJS)
78.157.59.0	78.157.59.255	Dade Samane Fanava Company (PJS)
78.157.60.0	78.157.61.255	Dade Samane Fanava Company (PJS)
78.157.62.0	78.157.62.255	Dade Samane Fanava Company (PJS)
78.157.63.0	78.157.63.255	Dade Samane Fanava Company (PJS)
78.158.160.0	78.158.160.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
78.158.161.0	78.158.161.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
78.158.162.0	78.158.162.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
78.158.163.0	78.158.163.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
78.158.164.0	78.158.164.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
78.158.165.0	78.158.165.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
78.158.166.0	78.158.166.255	Arvin Data Transfer Co. Ltd.
78.158.167.0	78.158.167.255	Khatam Academic Institute
78.158.168.0	78.158.168.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
78.158.169.0	78.158.169.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
78.158.170.0	78.158.171.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
78.158.172.0	78.158.175.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
78.158.176.0	78.158.176.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
78.158.177.0	78.158.177.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
78.158.178.0	78.158.179.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
78.158.180.0	78.158.180.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
78.158.181.0	78.158.181.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
78.158.182.0	78.158.183.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
78.158.184.0	78.158.184.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
78.158.185.0	78.158.185.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
78.158.186.0	78.158.187.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
78.158.188.0	78.158.191.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
79.127.0.0	79.127.3.255	Asiatech Data Transmission Company
79.127.4.0	79.127.7.255	Asiatech Data Transmission Company
79.127.8.0	79.127.11.255	Asiatech Data Transmission Company
79.127.12.0	79.127.15.255	Asiatech Data Transmission Company
79.127.16.0	79.127.19.255	Asiatech Data Transmission Company
79.127.20.0	79.127.21.255	Asiatech Data Transmission Company
79.127.22.0	79.127.22.255	Asiatech Data Transmission Company
79.127.23.0	79.127.23.255	Asiatech Data Transmission Company
79.127.24.0	79.127.24.255	Asiatech Data Transmission Company
79.127.25.0	79.127.25.255	Asiatech Data Transmission Company
79.127.26.0	79.127.26.255	Asiatech Data Transmission Company
79.127.27.0	79.127.27.255	Asiatech Data Transmission Company
79.127.28.0	79.127.31.255	Asiatech Data Transmission Company
79.127.32.0	79.127.39.255	Asiatech Data Transmission Company
79.127.40.0	79.127.41.255	Asiatech Data Transmission Company
79.127.42.0	79.127.42.255	Asiatech Data Transmission Company
79.127.43.0	79.127.43.255	Asiatech Data Transmission Company
79.127.44.0	79.127.45.255	Ertebatat Sabet Parsian Co. PJS
79.127.46.0	79.127.46.255	Secure Infrastructure of Transactional Services Company PJSC
79.127.47.0	79.127.47.255	Asiatech Data Transmission Company
79.127.48.0	79.127.55.255	Asiatech Data Transmission Company
79.127.56.0	79.127.57.255	Asiatech Data Transmission Company
79.127.58.0	79.127.58.255	Asiatech Data Transmission Company
79.127.59.0	79.127.59.255	Asiatech Data Transmission Company
79.127.60.0	79.127.61.255	Ertebatat Sabet Parsian Co. PJS
79.127.62.0	79.127.63.255	Asiatech Data Transmission Company
79.127.64.0	79.127.64.255	University of Sistan and Baluchestan
79.127.65.0	79.127.65.255	Institute for Research in Fundamental Sciences
79.127.66.0	79.127.67.255	Asiatech Data Transmission Company
79.127.68.0	79.127.69.255	Asiatech Data Transmission Company
79.127.70.0	79.127.70.255	Asiatech Data Transmission Company
79.127.71.0	79.127.71.255	Asiatech Data Transmission Company
79.127.72.0	79.127.72.255	Asiatech Data Transmission Company
79.127.73.0	79.127.73.255	Rayan Bours PJSC
79.127.74.0	79.127.75.255	Asiatech Data Transmission Company
79.127.76.0	79.127.76.255	Asiatech Data Transmission Company
79.127.77.0	79.127.77.255	Asiatech Data Transmission Company
79.127.78.0	79.127.79.255	Asiatech Data Transmission Company
79.127.80.0	79.127.83.255	Asiatech Data Transmission Company
79.127.84.0	79.127.87.255	Asiatech Data Transmission Company
79.127.88.0	79.127.95.255	Asiatech Data Transmission Company
79.127.96.0	79.127.111.255	Asiatech Data Transmission Company
79.127.112.0	79.127.113.255	Asiatech Data Transmission Company
79.127.114.0	79.127.115.255	Asiatech Data Transmission Company
79.127.116.0	79.127.119.255	Asiatech Data Transmission Company
79.127.120.0	79.127.121.255	Asiatech Data Transmission Company
79.127.122.0	79.127.123.255	Asiatech Data Transmission Company
79.127.124.0	79.127.124.255	Asiatech Data Transmission Company
79.127.125.0	79.127.125.255	Asiatech Data Transmission Company
79.127.126.0	79.127.126.255	Asiatech Data Transmission Company
79.127.127.0	79.127.127.255	Asiatech Data Transmission Company
79.132.192.0	79.132.192.255	Morva System Engineering Technical Company Private Joint Stock
79.132.193.0	79.132.193.255	Morva System Engineering Technical Company Private Joint Stock
79.132.200.0	79.132.203.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
79.132.204.0	79.132.204.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
79.132.205.0	79.132.205.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
79.132.206.0	79.132.207.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
79.132.208.0	79.132.209.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
79.132.210.0	79.132.210.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
79.132.211.0	79.132.211.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
79.132.212.0	79.132.212.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
79.132.213.0	79.132.213.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
79.132.214.0	79.132.214.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
79.132.215.0	79.132.215.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
79.132.216.0	79.132.216.255	Khatam Academic Institute
79.132.217.0	79.132.217.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
79.132.218.0	79.132.219.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
79.132.220.0	79.132.221.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
79.132.222.0	79.132.223.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
79.143.84.0	79.143.84.255	Dade Pardazi Mobinhost Co Ltd
79.143.85.0	79.143.85.255	FanAvaran Mihan Mizban PJSC
79.143.86.0	79.143.86.255	Mohammad Sheikhe Sajadie
79.174.160.0	79.174.163.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
79.174.164.0	79.174.164.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
79.174.165.0	79.174.165.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
79.174.166.0	79.174.167.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
79.175.128.0	79.175.128.255	Afranet
79.175.129.0	79.175.129.255	Afranet
79.175.130.0	79.175.130.255	Afranet
79.175.131.0	79.175.131.255	Afranet
79.175.132.0	79.175.132.255	Afranet
79.175.133.0	79.175.133.255	Afranet
79.175.134.0	79.175.134.255	Afranet
79.175.135.0	79.175.135.255	Afranet
79.175.136.0	79.175.136.255	Afranet
79.175.137.0	79.175.137.255	Afranet
79.175.138.0	79.175.138.255	Afranet
79.175.139.0	79.175.139.255	Afranet
79.175.140.0	79.175.140.255	Afranet
79.175.141.0	79.175.141.255	Afranet
79.175.142.0	79.175.142.255	Afranet
79.175.143.0	79.175.143.255	Afranet
79.175.144.0	79.175.144.255	Afranet
79.175.145.0	79.175.145.255	Afranet
79.175.146.0	79.175.146.255	Afranet
79.175.147.0	79.175.147.255	Afranet
79.175.148.0	79.175.148.255	Afranet
79.175.149.0	79.175.149.255	Afranet
79.175.150.0	79.175.150.255	Afranet
79.175.151.0	79.175.151.255	Afranet
79.175.152.0	79.175.152.255	Afranet
79.175.153.0	79.175.153.255	Afranet
79.175.154.0	79.175.154.255	Afranet
79.175.155.0	79.175.155.255	Afranet
79.175.156.0	79.175.156.255	Afranet
79.175.157.0	79.175.157.255	Afranet
79.175.158.0	79.175.158.255	Afranet
79.175.159.0	79.175.159.255	Afranet
79.175.160.0	79.175.160.255	Afranet
79.175.161.0	79.175.161.255	Afranet
79.175.162.0	79.175.162.255	Afranet
79.175.163.0	79.175.163.255	Afranet
79.175.164.0	79.175.164.255	Afranet
79.175.165.0	79.175.165.255	Afranet
79.175.166.0	79.175.166.255	Afranet
79.175.167.0	79.175.167.255	Afranet
79.175.168.0	79.175.168.255	Afranet
79.175.169.0	79.175.169.255	Afranet
79.175.170.0	79.175.170.255	Afranet
79.175.171.0	79.175.171.255	Afranet
79.175.172.0	79.175.172.255	Afranet
79.175.173.0	79.175.173.255	Afranet
79.175.174.0	79.175.174.255	Afranet
79.175.175.0	79.175.175.255	Afranet
79.175.176.0	79.175.176.255	Afranet
79.175.177.0	79.175.177.255	Afranet
79.175.178.0	79.175.178.255	Afranet
79.175.179.0	79.175.179.255	Afranet
79.175.180.0	79.175.180.255	Afranet
79.175.181.0	79.175.181.255	Afranet
79.175.182.0	79.175.182.255	Afranet
79.175.183.0	79.175.183.255	Afranet
79.175.184.0	79.175.184.255	Afranet
79.175.185.0	79.175.185.255	Afranet
79.175.186.0	79.175.186.255	Afranet
79.175.187.0	79.175.187.255	Afranet
79.175.188.0	79.175.188.255	Afranet
79.175.189.0	79.175.189.255	Afranet
79.175.190.0	79.175.190.255	Afranet
79.175.191.0	79.175.191.255	Afranet
80.66.176.0	80.66.177.255	University of Tehran
80.66.178.0	80.66.179.255	University of Tehran
80.66.180.0	80.66.183.255	University of Tehran
80.66.184.0	80.66.191.255	University of Tehran
80.71.112.0	80.71.112.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
80.71.113.0	80.71.113.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
80.71.114.0	80.71.114.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
80.71.115.0	80.71.115.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
80.71.116.0	80.71.119.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
80.71.120.0	80.71.120.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
80.71.121.0	80.71.121.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
80.71.122.0	80.71.123.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
80.71.124.0	80.71.125.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
80.71.126.0	80.71.126.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
80.71.127.0	80.71.127.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
80.71.149.0	80.71.149.255	Sefroyek Pardaz Engineering PJSC
80.75.0.0	80.75.0.255	Afranet
80.75.1.0	80.75.1.255	Afranet
80.75.2.0	80.75.2.255	Afranet
80.75.3.0	80.75.3.255	Afranet
80.75.4.0	80.75.4.255	Afranet
80.75.5.0	80.75.5.255	Afranet
80.75.6.0	80.75.6.255	Afranet
80.75.7.0	80.75.7.255	Afranet
80.75.8.0	80.75.8.255	Afranet
80.75.9.0	80.75.9.255	Afranet
80.75.10.0	80.75.10.255	Afranet
80.75.11.0	80.75.11.255	Afranet
80.75.12.0	80.75.12.255	Afranet
80.75.13.0	80.75.13.255	Afranet
80.75.14.0	80.75.14.255	Afranet
80.75.15.0	80.75.15.255	Afranet
80.75.213.0	80.75.213.255	Mobin Arvand Infrastructure Company Ltd
80.91.208.0	80.91.208.255	Noavaran Shabakeh Sabz Mehregan (Ltd.)
80.94.80.0	80.94.81.255	Mobin Net Communication Company (Private Joint Stock)
80.191.0.0	80.191.0.255	Iran Telecommunication Company Pjs
80.191.1.0	80.191.1.255	Iran Telecommunication Company Pjs
80.191.2.0	80.191.3.255	Iran Telecommunication Company Pjs
80.191.4.0	80.191.7.255	Iran Telecommunication Company Pjs
80.191.8.0	80.191.8.255	Iran Telecommunication Company Pjs
80.191.9.0	80.191.9.255	Iran Telecommunication Company Pjs
80.191.10.0	80.191.11.255	Iran Telecommunication Company Pjs
80.191.12.0	80.191.15.255	Iran Telecommunication Company Pjs
80.191.16.0	80.191.17.255	Iran Telecommunication Company Pjs
80.191.18.0	80.191.18.255	Iran Telecommunication Company Pjs
80.191.19.0	80.191.19.255	Iran Telecommunication Company Pjs
80.191.20.0	80.191.20.255	Iran Telecommunication Company Pjs
80.191.21.0	80.191.21.255	Iran Telecommunication Company Pjs
80.191.22.0	80.191.22.255	Iran Telecommunication Company Pjs
80.191.23.0	80.191.23.255	Iran Telecommunication Company Pjs
80.191.24.0	80.191.24.255	Iran Information Technology Company PJSC
80.191.25.0	80.191.25.255	Iran Telecommunication Company Pjs
80.191.26.0	80.191.27.255	Iran Telecommunication Company Pjs
80.191.28.0	80.191.31.255	Iran Telecommunication Company Pjs
80.191.32.0	80.191.39.255	Iran Telecommunication Company Pjs
80.191.40.0	80.191.47.255	Iran Telecommunication Company Pjs
80.191.48.0	80.191.55.255	Iran Telecommunication Company Pjs
80.191.56.0	80.191.56.255	Telecommunication Infrastructure Company
80.191.57.0	80.191.57.255	Iran Telecommunication Company Pjs
80.191.58.0	80.191.59.255	Iran Telecommunication Company Pjs
80.191.60.0	80.191.60.255	Iran Telecommunication Company Pjs
80.191.61.0	80.191.61.255	Iran Telecommunication Company Pjs
80.191.62.0	80.191.63.255	Iran Telecommunication Company Pjs
80.191.64.0	80.191.71.255	Iran Telecommunication Company Pjs
80.191.72.0	80.191.73.255	Iran Telecommunication Company Pjs
80.191.74.0	80.191.74.255	Iran Telecommunication Company Pjs
80.191.75.0	80.191.75.255	Iran Telecommunication Company Pjs
80.191.76.0	80.191.79.255	Iran Telecommunication Company Pjs
80.191.80.0	80.191.80.255	Iran Telecommunication Company Pjs
80.191.81.0	80.191.81.255	Telecommunication Infrastructure Company
80.191.82.0	80.191.82.255	Iran Telecommunication Company Pjs
80.191.83.0	80.191.83.255	Iran Telecommunication Company Pjs
80.191.84.0	80.191.87.255	Iran Telecommunication Company Pjs
80.191.88.0	80.191.89.255	Iran Telecommunication Company Pjs
80.191.90.0	80.191.90.255	Iran Telecommunication Company Pjs
80.191.91.0	80.191.91.255	Iran Telecommunication Company Pjs
80.191.92.0	80.191.95.255	Iran Telecommunication Company Pjs
80.191.96.0	80.191.99.255	Iran Telecommunication Company Pjs
80.191.100.0	80.191.100.255	Iran Telecommunication Company Pjs
80.191.101.0	80.191.101.255	Iran Telecommunication Company Pjs
80.191.102.0	80.191.102.255	Iran Telecommunication Company Pjs
80.191.103.0	80.191.103.255	Iran Telecommunication Company Pjs
80.191.104.0	80.191.107.255	Iran Telecommunication Company Pjs
80.191.108.0	80.191.108.255	Iran Telecommunication Company Pjs
80.191.109.0	80.191.109.255	Iran Telecommunication Company Pjs
80.191.110.0	80.191.111.255	Iran Telecommunication Company Pjs
80.191.112.0	80.191.127.255	Iran Telecommunication Company Pjs
80.191.128.0	80.191.135.255	Iran Telecommunication Company Pjs
80.191.136.0	80.191.137.255	Iran Telecommunication Company Pjs
80.191.138.0	80.191.139.255	Iran Telecommunication Company Pjs
80.191.140.0	80.191.143.255	Iran Telecommunication Company Pjs
80.191.144.0	80.191.151.255	Iran Telecommunication Company Pjs
80.191.152.0	80.191.159.255	Iran Telecommunication Company Pjs
80.191.160.0	80.191.167.255	Iran Telecommunication Company Pjs
80.191.168.0	80.191.171.255	Iran Telecommunication Company Pjs
80.191.172.0	80.191.173.255	University of Sistan and Baluchestan
80.191.174.0	80.191.175.255	Iran Telecommunication Company Pjs
80.191.176.0	80.191.177.255	Iran Telecommunication Company Pjs
80.191.178.0	80.191.178.255	Iran Telecommunication Company Pjs
80.191.179.0	80.191.179.255	Iran Telecommunication Company Pjs
80.191.180.0	80.191.181.255	Iran Telecommunication Company Pjs
80.191.182.0	80.191.183.255	Iran Telecommunication Company Pjs
80.191.184.0	80.191.185.255	Iran Telecommunication Company Pjs
80.191.186.0	80.191.187.255	Iran Telecommunication Company Pjs
80.191.188.0	80.191.191.255	Iran Telecommunication Company Pjs
80.191.192.0	80.191.193.255	Iran Telecommunication Company Pjs
80.191.194.0	80.191.195.255	Iran Telecommunication Company Pjs
80.191.196.0	80.191.196.255	Iran Telecommunication Company Pjs
80.191.197.0	80.191.197.255	Iran Telecommunication Company Pjs
80.191.198.0	80.191.198.255	Iran Telecommunication Company Pjs
80.191.199.0	80.191.199.255	Iran Telecommunication Company Pjs
80.191.200.0	80.191.200.255	Iran Telecommunication Company Pjs
80.191.201.0	80.191.201.255	Iran Telecommunication Company Pjs
80.191.202.0	80.191.203.255	Iran Telecommunication Company Pjs
80.191.204.0	80.191.207.255	Iran Telecommunication Company Pjs
80.191.208.0	80.191.209.255	Iran Telecommunication Company Pjs
80.191.210.0	80.191.210.255	Iran Telecommunication Company Pjs
80.191.211.0	80.191.211.255	Iran Telecommunication Company Pjs
80.191.212.0	80.191.215.255	Iran Telecommunication Company Pjs
80.191.216.0	80.191.219.255	Iran Telecommunication Company Pjs
80.191.220.0	80.191.223.255	Iran Telecommunication Company Pjs
80.191.224.0	80.191.231.255	Iran Telecommunication Company Pjs
80.191.232.0	80.191.235.255	Iran Telecommunication Company Pjs
80.191.236.0	80.191.239.255	Iran Telecommunication Company Pjs
80.191.240.0	80.191.247.255	Iran Telecommunication Company Pjs
80.191.248.0	80.191.251.255	Iran Telecommunication Company Pjs
80.191.252.0	80.191.255.255	Iran Telecommunication Company Pjs
80.210.0.0	80.210.15.255	Iran Telecommunication Company Pjs
80.210.16.0	80.210.16.255	Iran Telecommunication Company Pjs
80.210.17.0	80.210.17.255	Iran Telecommunication Company Pjs
80.210.18.0	80.210.19.255	Iran Telecommunication Company Pjs
80.210.20.0	80.210.23.255	Iran Telecommunication Company Pjs
80.210.24.0	80.210.31.255	Iran Telecommunication Company Pjs
80.210.32.0	80.210.63.255	Iran Telecommunication Company Pjs
80.210.128.0	80.210.143.255	Iran Telecommunication Company Pjs
80.210.144.0	80.210.159.255	Iran Telecommunication Company Pjs
80.210.160.0	80.210.175.255	Iran Telecommunication Company Pjs
80.210.176.0	80.210.191.255	Iran Telecommunication Company Pjs
80.210.192.0	80.210.207.255	Iran Telecommunication Company Pjs
80.210.208.0	80.210.223.255	Iran Telecommunication Company Pjs
80.210.224.0	80.210.239.255	Iran Telecommunication Company Pjs
80.210.240.0	80.210.247.255	Iran Telecommunication Company Pjs
80.210.248.0	80.210.255.255	Iran Telecommunication Company Pjs
80.242.0.0	80.242.15.255	Mobile Communication Company of Iran Plc
80.244.7.0	80.244.7.255	Web Dadeh Paydar Co (Ltd)
80.244.11.0	80.244.11.255	Atis Omran Sevin PSJ
80.249.112.0	80.249.113.255	Sefroyek Pardaz Engineering PJSC
80.249.114.0	80.249.114.255	Sefroyek Pardaz Engineering PJSC
80.249.115.0	80.249.115.255	Netafraz Iranian Ltd.
80.250.192.0	80.250.195.255	Iran Telecommunication Company Pjs
80.250.196.0	80.250.199.255	Iran Telecommunication Company Pjs
80.250.200.0	80.250.207.255	Iran Telecommunication Company Pjs
80.253.128.0	80.253.128.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
80.253.129.0	80.253.129.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
80.253.130.0	80.253.131.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
80.253.132.0	80.253.132.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
80.253.133.0	80.253.133.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
80.253.134.0	80.253.134.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
80.253.135.0	80.253.135.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
80.253.136.0	80.253.136.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
80.253.137.0	80.253.137.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
80.253.138.0	80.253.138.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
80.253.139.0	80.253.139.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
80.253.140.0	80.253.140.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
80.253.141.0	80.253.141.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
80.253.142.0	80.253.143.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
80.253.144.0	80.253.144.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
80.253.145.0	80.253.145.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
80.253.146.0	80.253.146.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
80.253.147.0	80.253.147.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
80.253.148.0	80.253.148.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
80.253.149.0	80.253.149.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
80.253.150.0	80.253.150.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
80.253.151.0	80.253.151.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
80.253.152.0	80.253.152.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
80.253.153.0	80.253.153.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
80.253.154.0	80.253.155.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
80.253.156.0	80.253.156.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
80.253.157.0	80.253.157.255	Rayan Hamafza Company Public Joint Stock
80.253.158.0	80.253.158.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
80.253.159.0	80.253.159.255	Naji Research & Development Company Private Join-Stock
81.12.0.0	81.12.0.255	Respina Networks & Beyond PJSC
81.12.1.0	81.12.1.255	Respina Networks & Beyond PJSC
81.12.2.0	81.12.2.255	Respina Networks & Beyond PJSC
81.12.3.0	81.12.3.255	Respina Networks & Beyond PJSC
81.12.4.0	81.12.4.255	Respina Networks & Beyond PJSC
81.12.5.0	81.12.5.255	Respina Networks & Beyond PJSC
81.12.6.0	81.12.6.255	Respina Networks & Beyond PJSC
81.12.7.0	81.12.7.255	Respina Networks & Beyond PJSC
81.12.8.0	81.12.11.255	Soroush Rasanheh Company Ltd
81.12.12.0	81.12.12.255	Institute for Research in Fundamental Sciences
81.12.13.0	81.12.13.255	Soroush Rasanheh Company Ltd
81.12.14.0	81.12.15.255	Soroush Rasanheh Company Ltd
81.12.16.0	81.12.23.255	Soroush Rasanheh Company Ltd
81.12.24.0	81.12.24.255	Dade Pardazi Mobinhost Co Ltd
81.12.25.0	81.12.25.255	Dade Pardazi Mobinhost Co Ltd
81.12.26.0	81.12.26.255	Dade Pardazi Mobinhost Co Ltd
81.12.27.0	81.12.27.255	Dade Pardazi Mobinhost Co Ltd
81.12.28.0	81.12.28.255	Iran Telecommunication Company Pjs
81.12.29.0	81.12.29.255	Alborz Insurance Company
81.12.30.0	81.12.30.255	Dade Pardazi Mobinhost Co Ltd
81.12.31.0	81.12.31.255	Dade Pardazi Mobinhost Co Ltd
81.12.32.0	81.12.32.255	Pars Abr Toseeh Ertebatat Ltd
81.12.33.0	81.12.33.255	Pars Abr Toseeh Ertebatat Ltd
81.12.34.0	81.12.34.255	Pars Abr Toseeh Ertebatat Ltd
81.12.35.0	81.12.35.255	Pars Abr Toseeh Ertebatat Ltd
81.12.36.0	81.12.36.255	Respina Networks & Beyond PJSC
81.12.37.0	81.12.37.255	Respina Networks & Beyond PJSC
81.12.38.0	81.12.38.255	Respina Networks & Beyond PJSC
81.12.39.0	81.12.39.255	Dade Pardaz Kimia Pouyesh PJS.
81.12.40.0	81.12.40.255	Respina Networks & Beyond PJSC
81.12.41.0	81.12.41.255	Respina Networks & Beyond PJSC
81.12.42.0	81.12.42.255	Respina Networks & Beyond PJSC
81.12.43.0	81.12.43.255	Respina Networks & Beyond PJSC
81.12.44.0	81.12.44.255	Respina Networks & Beyond PJSC
81.12.45.0	81.12.45.255	Respina Networks & Beyond PJSC
81.12.46.0	81.12.46.255	Faratechnology
81.12.47.0	81.12.47.255	Respina Networks & Beyond PJSC
81.12.48.0	81.12.49.255	Soroush Rasanheh Company Ltd
81.12.50.0	81.12.51.255	Soroush Rasanheh Company Ltd
81.12.52.0	81.12.52.255	Pars Shabakeh Azarakhsh LLC
81.12.53.0	81.12.53.255	Pars Shabakeh Azarakhsh LLC
81.12.54.0	81.12.54.255	FanAvaran Mihan Mizban PJSC
81.12.55.0	81.12.55.255	Respina Networks & Beyond PJSC
81.12.56.0	81.12.56.255	Respina Networks & Beyond PJSC
81.12.57.0	81.12.57.255	Arsan Technology Development Co Ltd
81.12.58.0	81.12.58.255	Soroush Rasanheh Company Ltd
81.12.59.0	81.12.59.255	Soroush Rasanheh Company Ltd
81.12.60.0	81.12.60.255	Respina Networks & Beyond PJSC
81.12.61.0	81.12.61.255	Respina Networks & Beyond PJSC
81.12.62.0	81.12.62.255	Respina Networks & Beyond PJSC
81.12.63.0	81.12.63.255	Respina Networks & Beyond PJSC
81.12.64.0	81.12.64.255	Respina Networks & Beyond PJSC
81.12.65.0	81.12.65.255	Respina Networks & Beyond PJSC
81.12.66.0	81.12.66.255	Respina Networks & Beyond PJSC
81.12.67.0	81.12.67.255	Respina Networks & Beyond PJSC
81.12.68.0	81.12.68.255	Respina Networks & Beyond PJSC
81.12.69.0	81.12.69.255	Respina Networks & Beyond PJSC
81.12.70.0	81.12.70.255	Respina Networks & Beyond PJSC
81.12.71.0	81.12.71.255	Respina Networks & Beyond PJSC
81.12.72.0	81.12.72.255	Respina Networks & Beyond PJSC
81.12.73.0	81.12.73.255	Respina Networks & Beyond PJSC
81.12.74.0	81.12.74.255	Respina Networks & Beyond PJSC
81.12.75.0	81.12.75.255	Respina Networks & Beyond PJSC
81.12.76.0	81.12.76.255	Respina Networks & Beyond PJSC
81.12.77.0	81.12.77.255	Respina Networks & Beyond PJSC
81.12.78.0	81.12.78.255	Fara Net Gostar Kavoshgaran Aftab
81.12.79.0	81.12.79.255	Respina Networks & Beyond PJSC
81.12.80.0	81.12.80.255	Pars Abr Toseeh Ertebatat Ltd
81.12.81.0	81.12.81.255	Respina Networks & Beyond PJSC
81.12.82.0	81.12.83.255	Respina Networks & Beyond PJSC
81.12.84.0	81.12.85.255	Respina Networks & Beyond PJSC
81.12.86.0	81.12.86.255	Respina Networks & Beyond PJSC
81.12.87.0	81.12.87.255	Respina Networks & Beyond PJSC
81.12.88.0	81.12.88.255	Respina Networks & Beyond PJSC
81.12.89.0	81.12.89.255	Respina Networks & Beyond PJSC
81.12.90.0	81.12.91.255	Respina Networks & Beyond PJSC
81.12.92.0	81.12.92.255	Arsan Technology Development Co Ltd
81.12.93.0	81.12.93.255	Respina Networks & Beyond PJSC
81.12.94.0	81.12.94.255	Respina Networks & Beyond PJSC
81.12.95.0	81.12.95.255	Respina Networks & Beyond PJSC
81.12.96.0	81.12.96.255	Respina Networks & Beyond PJSC
81.12.97.0	81.12.97.255	Respina Networks & Beyond PJSC
81.12.98.0	81.12.98.255	Respina Networks & Beyond PJSC
81.12.99.0	81.12.99.255	Respina Networks & Beyond PJSC
81.12.100.0	81.12.100.255	Respina Networks & Beyond PJSC
81.12.101.0	81.12.101.255	Respina Networks & Beyond PJSC
81.12.102.0	81.12.102.255	Respina Networks & Beyond PJSC
81.12.103.0	81.12.103.255	Respina Networks & Beyond PJSC
81.12.104.0	81.12.104.255	Respina Networks & Beyond PJSC
81.12.105.0	81.12.105.255	Respina Networks & Beyond PJSC
81.12.106.0	81.12.106.255	Respina Networks & Beyond PJSC
81.12.107.0	81.12.107.255	Respina Networks & Beyond PJSC
81.12.108.0	81.12.108.255	Respina Networks & Beyond PJSC
81.12.109.0	81.12.109.255	Respina Networks & Beyond PJSC
81.12.110.0	81.12.110.255	Respina Networks & Beyond PJSC
81.12.111.0	81.12.111.255	Respina Networks & Beyond PJSC
81.12.112.0	81.12.112.255	Respina Networks & Beyond PJSC
81.12.113.0	81.12.113.255	Respina Networks & Beyond PJSC
81.12.114.0	81.12.114.255	Respina Networks & Beyond PJSC
81.12.115.0	81.12.115.255	Respina Networks & Beyond PJSC
81.12.116.0	81.12.116.255	Respina Networks & Beyond PJSC
81.12.117.0	81.12.117.255	Respina Networks & Beyond PJSC
81.12.118.0	81.12.118.255	Respina Networks & Beyond PJSC
81.12.119.0	81.12.119.255	Respina Networks & Beyond PJSC
81.12.120.0	81.12.120.255	Respina Networks & Beyond PJSC
81.12.121.0	81.12.121.255	Respina Networks & Beyond PJSC
81.12.122.0	81.12.122.255	Respina Networks & Beyond PJSC
81.12.123.0	81.12.123.255	Respina Networks & Beyond PJSC
81.12.124.0	81.12.124.255	Respina Networks & Beyond PJSC
81.12.125.0	81.12.125.255	Respina Networks & Beyond PJSC
81.12.126.0	81.12.126.255	Respina Networks & Beyond PJSC
81.12.127.0	81.12.127.255	Respina Networks & Beyond PJSC
81.16.112.0	81.16.127.255	Iran Telecommunication Company Pjs
81.28.32.0	81.28.32.255	IsIran
81.28.33.0	81.28.33.255	IsIran
81.28.34.0	81.28.34.255	IsIran
81.28.35.0	81.28.35.255	IsIran
81.28.36.0	81.28.36.255	IsIran
81.28.37.0	81.28.37.255	IsIran
81.28.38.0	81.28.38.255	IsIran
81.28.39.0	81.28.39.255	IsIran
81.28.40.0	81.28.40.255	IsIran
81.28.41.0	81.28.41.255	IsIran
81.28.42.0	81.28.42.255	IsIran
81.28.43.0	81.28.43.255	IsIran
81.28.44.0	81.28.44.255	IsIran
81.28.45.0	81.28.45.255	IsIran
81.28.46.0	81.28.46.255	IsIran
81.28.47.0	81.28.47.255	IsIran
81.28.48.0	81.28.48.255	IsIran
81.28.49.0	81.28.49.255	IsIran
81.28.50.0	81.28.50.255	IsIran
81.28.51.0	81.28.51.255	IsIran
81.28.52.0	81.28.52.255	IsIran
81.28.53.0	81.28.53.255	IsIran
81.28.54.0	81.28.54.255	IsIran
81.28.55.0	81.28.55.255	IsIran
81.28.56.0	81.28.56.255	IsIran
81.28.57.0	81.28.57.255	IsIran
81.28.58.0	81.28.58.255	IsIran
81.28.59.0	81.28.59.255	IsIran
81.28.60.0	81.28.60.255	IsIran
81.28.61.0	81.28.61.255	IsIran
81.28.62.0	81.28.63.255	IsIran
81.28.252.0	81.28.255.255	Gostaresh Ertebat Azin Kia Company PJSC
81.29.240.0	81.29.240.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
81.29.241.0	81.29.241.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
81.29.242.0	81.29.242.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
81.29.243.0	81.29.243.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
81.29.244.0	81.29.247.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
81.29.248.0	81.29.249.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
81.29.250.0	81.29.251.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
81.29.252.0	81.29.253.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
81.29.254.0	81.29.254.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
81.29.255.0	81.29.255.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
81.30.108.0	81.30.108.255	Web Dadeh Paydar Co (Ltd)
81.31.160.0	81.31.160.255	Sharif University of Technology
81.31.161.0	81.31.161.255	Sharif University of Technology
81.31.162.0	81.31.163.255	Sharif University of Technology
81.31.164.0	81.31.167.255	Sharif University of Technology
81.31.168.0	81.31.168.255	Sharif University of Technology
81.31.169.0	81.31.169.255	Sharif University of Technology
81.31.170.0	81.31.170.255	Sharif University of Technology
81.31.171.0	81.31.171.255	Sharif University of Technology
81.31.172.0	81.31.175.255	Sharif University of Technology
81.31.176.0	81.31.179.255	Sharif University of Technology
81.31.180.0	81.31.183.255	Sharif University of Technology
81.31.184.0	81.31.185.255	Sharif University of Technology
81.31.186.0	81.31.186.255	Sharif University of Technology
81.31.187.0	81.31.187.255	Sharif University of Technology
81.31.188.0	81.31.191.255	Sharif University of Technology
81.31.224.0	81.31.224.255	Chapar Rasaneh LLC
81.31.225.0	81.31.225.255	Chapar Rasaneh LLC
81.31.226.0	81.31.226.255	Chapar Rasaneh LLC
81.31.227.0	81.31.227.255	Chapar Rasaneh LLC
81.31.228.0	81.31.228.255	Chapar Rasaneh LLC
81.31.229.0	81.31.229.255	Chapar Rasaneh LLC
81.31.230.0	81.31.230.255	Chapar Rasaneh LLC
81.31.233.0	81.31.233.255	Chapar Rasaneh LLC
81.31.236.0	81.31.236.255	Chapar Rasaneh LLC
81.31.237.0	81.31.237.255	Chapar Rasaneh LLC
81.31.238.0	81.31.238.255	Chapar Rasaneh LLC
81.31.239.0	81.31.239.255	Chapar Rasaneh LLC
81.31.240.0	81.31.240.255	Chapar Rasaneh LLC
81.31.241.0	81.31.241.255	Chapar Rasaneh LLC
81.31.248.0	81.31.248.255	Chapar Rasaneh LLC
81.31.249.0	81.31.249.255	Chapar Rasaneh LLC
81.31.250.0	81.31.250.255	Chapar Rasaneh LLC
81.31.251.0	81.31.251.255	Chapar Rasaneh LLC
81.90.144.0	81.90.144.255	Afranet
81.90.145.0	81.90.145.255	Afranet
81.90.146.0	81.90.146.255	Afranet
81.90.147.0	81.90.147.255	Afranet
81.90.148.0	81.90.148.255	Afranet
81.90.149.0	81.90.149.255	Afranet
81.90.150.0	81.90.150.255	Afranet
81.90.151.0	81.90.151.255	Afranet
81.90.152.0	81.90.152.255	Afranet
81.90.153.0	81.90.153.255	Afranet
81.90.154.0	81.90.154.255	Afranet
81.90.155.0	81.90.155.255	Afranet
81.90.156.0	81.90.156.255	Afranet
81.90.157.0	81.90.157.255	Afranet
81.90.158.0	81.90.158.255	Afranet
81.90.159.0	81.90.159.255	Afranet
81.91.128.0	81.91.129.255	Datak Company LLC
81.91.130.0	81.91.130.255	Datak Company LLC
81.91.131.0	81.91.131.255	Datak Company LLC
81.91.132.0	81.91.135.255	Datak Company LLC
81.91.136.0	81.91.139.255	Datak Company LLC
81.91.140.0	81.91.143.255	Datak Company LLC
81.91.144.0	81.91.147.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
81.91.148.0	81.91.149.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
81.91.150.0	81.91.151.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
81.91.152.0	81.91.159.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
81.163.0.0	81.163.1.255	Rasana Pishtaz Iranian Service Cooperative Co.
81.163.2.0	81.163.3.255	Rasana Pishtaz Iranian Service Cooperative Co.
81.163.4.0	81.163.5.255	Rasana Pishtaz Iranian Service Cooperative Co.
81.163.6.0	81.163.7.255	Rasana Pishtaz Iranian Service Cooperative Co.
82.99.192.0	82.99.195.255	Parsan Lin Co. PJS
82.99.196.0	82.99.196.255	Parsan Lin Co. PJS
82.99.197.0	82.99.197.255	Parsan Lin Co. PJS
82.99.198.0	82.99.198.255	Parsan Lin Co. PJS
82.99.199.0	82.99.199.255	Parsan Lin Co. PJS
82.99.200.0	82.99.200.255	Parsan Lin Co. PJS
82.99.201.0	82.99.201.255	Parsan Lin Co. PJS
82.99.202.0	82.99.203.255	Parsan Lin Co. PJS
82.99.204.0	82.99.207.255	Parsan Lin Co. PJS
82.99.208.0	82.99.211.255	Parsan Lin Co. PJS
82.99.212.0	82.99.213.255	Parsan Lin Co. PJS
82.99.214.0	82.99.214.255	Parsan Lin Co. PJS
82.99.215.0	82.99.215.255	Parsan Lin Co. PJS
82.99.216.0	82.99.223.255	Parsan Lin Co. PJS
82.99.224.0	82.99.224.255	Tejarat Iranian Technology Infrastructure PJSC
82.99.225.0	82.99.225.255	Parsan Lin Co. PJS
82.99.226.0	82.99.227.255	Parsan Lin Co. PJS
82.99.228.0	82.99.228.255	Parsan Lin Co. PJS
82.99.229.0	82.99.229.255	Parsan Lin Co. PJS
82.99.230.0	82.99.231.255	Parsan Lin Co. PJS
82.99.232.0	82.99.233.255	Parsan Lin Co. PJS
82.99.234.0	82.99.234.255	Parsan Lin Co. PJS
82.99.235.0	82.99.235.255	Parsan Lin Co. PJS
82.99.236.0	82.99.239.255	Parsan Lin Co. PJS
82.99.240.0	82.99.247.255	Parsan Lin Co. PJS
82.99.248.0	82.99.251.255	Parsan Lin Co. PJS
82.99.252.0	82.99.252.255	Parsan Lin Co. PJS
82.99.253.0	82.99.253.255	Parsan Lin Co. PJS
82.99.254.0	82.99.255.255	Parsan Lin Co. PJS
82.102.9.0	82.102.9.255	Alma Fan Avaran Dehkade Jahani Company Ltd
82.102.14.0	82.102.14.255	Alma Fan Avaran Dehkade Jahani Company Ltd
82.102.15.0	82.102.15.255	Alma Fan Avaran Dehkade Jahani Company Ltd
82.180.192.0	82.180.192.255	Mobile Communication Company of Iran Plc
82.180.193.0	82.180.193.255	Mobile Communication Company of Iran Plc
82.180.194.0	82.180.195.255	Mobile Communication Company of Iran Plc
82.180.196.0	82.180.199.255	Mobile Communication Company of Iran Plc
82.180.200.0	82.180.207.255	Mobile Communication Company of Iran Plc
82.180.208.0	82.180.223.255	Mobile Communication Company of Iran Plc
82.180.224.0	82.180.239.255	Mobile Communication Company of Iran Plc
82.180.240.0	82.180.255.255	Mobile Communication Company of Iran Plc
83.97.72.0	83.97.72.255	Fekr Pardaz Shomal PLC
83.120.0.0	83.120.0.255	Mobile Communication Company of Iran Plc
83.120.1.0	83.120.1.255	Mobile Communication Company of Iran Plc
83.120.2.0	83.120.3.255	Mobile Communication Company of Iran Plc
83.120.4.0	83.120.7.255	Mobile Communication Company of Iran Plc
83.120.8.0	83.120.15.255	Mobile Communication Company of Iran Plc
83.120.16.0	83.120.31.255	Mobile Communication Company of Iran Plc
83.120.32.0	83.120.47.255	Mobile Communication Company of Iran Plc
83.120.48.0	83.120.63.255	Mobile Communication Company of Iran Plc
83.120.64.0	83.120.64.255	Mobile Communication Company of Iran Plc
83.120.65.0	83.120.65.255	Mobile Communication Company of Iran Plc
83.120.66.0	83.120.67.255	Mobile Communication Company of Iran Plc
83.120.68.0	83.120.71.255	Mobile Communication Company of Iran Plc
83.120.72.0	83.120.79.255	Mobile Communication Company of Iran Plc
83.120.80.0	83.120.95.255	Mobile Communication Company of Iran Plc
83.120.96.0	83.120.111.255	Mobile Communication Company of Iran Plc
83.120.112.0	83.120.127.255	Mobile Communication Company of Iran Plc
83.120.128.0	83.120.128.255	Mobile Communication Company of Iran Plc
83.120.129.0	83.120.129.255	Mobile Communication Company of Iran Plc
83.120.130.0	83.120.131.255	Mobile Communication Company of Iran Plc
83.120.132.0	83.120.135.255	Mobile Communication Company of Iran Plc
83.120.136.0	83.120.143.255	Mobile Communication Company of Iran Plc
83.120.144.0	83.120.159.255	Mobile Communication Company of Iran Plc
83.120.160.0	83.120.175.255	Mobile Communication Company of Iran Plc
83.120.176.0	83.120.191.255	Mobile Communication Company of Iran Plc
83.120.192.0	83.120.207.255	Mobile Communication Company of Iran Plc
83.120.208.0	83.120.223.255	Mobile Communication Company of Iran Plc
83.120.224.0	83.120.239.255	Mobile Communication Company of Iran Plc
83.120.240.0	83.120.255.255	Mobile Communication Company of Iran Plc
83.121.0.0	83.121.15.255	Mobile Communication Company of Iran Plc
83.121.16.0	83.121.31.255	Mobile Communication Company of Iran Plc
83.121.32.0	83.121.47.255	Mobile Communication Company of Iran Plc
83.121.48.0	83.121.63.255	Mobile Communication Company of Iran Plc
83.121.64.0	83.121.79.255	Mobile Communication Company of Iran Plc
83.121.80.0	83.121.95.255	Mobile Communication Company of Iran Plc
83.121.96.0	83.121.111.255	Mobile Communication Company of Iran Plc
83.121.112.0	83.121.127.255	Mobile Communication Company of Iran Plc
83.121.128.0	83.121.143.255	Mobile Communication Company of Iran Plc
83.121.144.0	83.121.159.255	Mobile Communication Company of Iran Plc
83.121.160.0	83.121.175.255	Mobile Communication Company of Iran Plc
83.121.176.0	83.121.191.255	Mobile Communication Company of Iran Plc
83.121.192.0	83.121.207.255	Mobile Communication Company of Iran Plc
83.121.208.0	83.121.223.255	Mobile Communication Company of Iran Plc
83.121.224.0	83.121.239.255	Mobile Communication Company of Iran Plc
83.121.240.0	83.121.255.255	Mobile Communication Company of Iran Plc
83.122.0.0	83.122.0.255	Mobile Communication Company of Iran Plc
83.122.1.0	83.122.1.255	Mobile Communication Company of Iran Plc
83.122.2.0	83.122.3.255	Mobile Communication Company of Iran Plc
83.122.4.0	83.122.7.255	Mobile Communication Company of Iran Plc
83.122.8.0	83.122.15.255	Mobile Communication Company of Iran Plc
83.122.16.0	83.122.31.255	Mobile Communication Company of Iran Plc
83.122.32.0	83.122.47.255	Mobile Communication Company of Iran Plc
83.122.48.0	83.122.63.255	Mobile Communication Company of Iran Plc
83.122.64.0	83.122.64.255	Mobile Communication Company of Iran Plc
83.122.65.0	83.122.65.255	Mobile Communication Company of Iran Plc
83.122.66.0	83.122.67.255	Mobile Communication Company of Iran Plc
83.122.68.0	83.122.71.255	Mobile Communication Company of Iran Plc
83.122.72.0	83.122.79.255	Mobile Communication Company of Iran Plc
83.122.80.0	83.122.95.255	Mobile Communication Company of Iran Plc
83.122.96.0	83.122.111.255	Mobile Communication Company of Iran Plc
83.122.112.0	83.122.127.255	Mobile Communication Company of Iran Plc
83.122.128.0	83.122.128.255	Mobile Communication Company of Iran Plc
83.122.129.0	83.122.129.255	Mobile Communication Company of Iran Plc
83.122.130.0	83.122.131.255	Mobile Communication Company of Iran Plc
83.122.132.0	83.122.135.255	Mobile Communication Company of Iran Plc
83.122.136.0	83.122.143.255	Mobile Communication Company of Iran Plc
83.122.144.0	83.122.159.255	Mobile Communication Company of Iran Plc
83.122.160.0	83.122.175.255	Mobile Communication Company of Iran Plc
83.122.176.0	83.122.191.255	Mobile Communication Company of Iran Plc
83.122.192.0	83.122.192.255	Mobile Communication Company of Iran Plc
83.122.193.0	83.122.193.255	Mobile Communication Company of Iran Plc
83.122.194.0	83.122.195.255	Mobile Communication Company of Iran Plc
83.122.196.0	83.122.199.255	Mobile Communication Company of Iran Plc
83.122.200.0	83.122.207.255	Mobile Communication Company of Iran Plc
83.122.208.0	83.122.223.255	Mobile Communication Company of Iran Plc
83.122.224.0	83.122.239.255	Mobile Communication Company of Iran Plc
83.122.240.0	83.122.255.255	Mobile Communication Company of Iran Plc
83.123.0.0	83.123.0.255	Mobile Communication Company of Iran Plc
83.123.1.0	83.123.1.255	Mobile Communication Company of Iran Plc
83.123.2.0	83.123.3.255	Mobile Communication Company of Iran Plc
83.123.4.0	83.123.7.255	Mobile Communication Company of Iran Plc
83.123.8.0	83.123.15.255	Mobile Communication Company of Iran Plc
83.123.16.0	83.123.31.255	Mobile Communication Company of Iran Plc
83.123.32.0	83.123.47.255	Mobile Communication Company of Iran Plc
83.123.48.0	83.123.63.255	Mobile Communication Company of Iran Plc
83.123.64.0	83.123.79.255	Mobile Communication Company of Iran Plc
83.123.80.0	83.123.95.255	Mobile Communication Company of Iran Plc
83.123.96.0	83.123.111.255	Mobile Communication Company of Iran Plc
83.123.112.0	83.123.127.255	Mobile Communication Company of Iran Plc
83.123.128.0	83.123.143.255	Mobile Communication Company of Iran Plc
83.123.144.0	83.123.159.255	Mobile Communication Company of Iran Plc
83.123.160.0	83.123.175.255	Mobile Communication Company of Iran Plc
83.123.176.0	83.123.191.255	Mobile Communication Company of Iran Plc
83.123.192.0	83.123.207.255	Mobile Communication Company of Iran Plc
83.123.208.0	83.123.223.255	Mobile Communication Company of Iran Plc
83.123.224.0	83.123.239.255	Mobile Communication Company of Iran Plc
83.123.240.0	83.123.255.255	Mobile Communication Company of Iran Plc
83.150.192.0	83.150.193.255	Sepehr Net Iranian PJSC
83.150.194.0	83.150.195.255	Sepehr Net Iranian PJSC
84.47.192.0	84.47.207.255	Iran Telecommunication Company Pjs
84.47.208.0	84.47.209.255	Iran Telecommunication Company Pjs
84.47.210.0	84.47.210.255	Iran Telecommunication Company Pjs
84.47.211.0	84.47.211.255	Iran Telecommunication Company Pjs
84.47.212.0	84.47.215.255	Iran Telecommunication Company Pjs
84.47.216.0	84.47.223.255	Iran Telecommunication Company Pjs
84.47.224.0	84.47.224.255	Sefroyek Pardaz Engineering PJSC
84.47.225.0	84.47.225.255	Sefroyek Pardaz Engineering PJSC
84.47.226.0	84.47.226.255	Sefroyek Pardaz Engineering PJSC
84.47.227.0	84.47.227.255	Sefroyek Pardaz Engineering PJSC
84.47.228.0	84.47.228.255	Sefroyek Pardaz Engineering PJSC
84.47.229.0	84.47.229.255	Sefroyek Pardaz Engineering PJSC
84.47.230.0	84.47.230.255	Sefroyek Pardaz Engineering PJSC
84.47.231.0	84.47.231.255	Sefroyek Pardaz Engineering PJSC
84.47.232.0	84.47.235.255	Iran Telecommunication Company Pjs
84.47.236.0	84.47.236.255	Iran Telecommunication Company Pjs
84.47.237.0	84.47.237.255	Iran Telecommunication Company Pjs
84.47.238.0	84.47.239.255	Iran Telecommunication Company Pjs
84.47.240.0	84.47.247.255	Iran Telecommunication Company Pjs
84.47.248.0	84.47.251.255	Iran Telecommunication Company Pjs
84.47.252.0	84.47.252.255	Iran Telecommunication Company Pjs
84.47.253.0	84.47.253.255	Iran Telecommunication Company Pjs
84.47.254.0	84.47.255.255	Iran Telecommunication Company Pjs
84.241.0.0	84.241.3.255	Aria Shatel PJSC
84.241.4.0	84.241.4.255	Aria Shatel PJSC
84.241.5.0	84.241.5.255	Aria Shatel PJSC
84.241.6.0	84.241.7.255	Aria Shatel PJSC
84.241.8.0	84.241.15.255	Aria Shatel PJSC
84.241.16.0	84.241.19.255	Aria Shatel PJSC
84.241.20.0	84.241.21.255	Aria Shatel PJSC
84.241.22.0	84.241.22.255	Aria Shatel PJSC
84.241.23.0	84.241.23.255	Aria Shatel PJSC
84.241.24.0	84.241.31.255	Aria Shatel PJSC
84.241.32.0	84.241.33.255	Aria Shatel PJSC
84.241.34.0	84.241.34.255	Aria Shatel PJSC
84.241.35.0	84.241.35.255	Aria Shatel PJSC
84.241.36.0	84.241.39.255	Aria Shatel PJSC
84.241.40.0	84.241.47.255	Aria Shatel PJSC
84.241.48.0	84.241.48.255	Aria Shatel PJSC
84.241.49.0	84.241.49.255	Aria Shatel PJSC
84.241.50.0	84.241.51.255	Aria Shatel PJSC
84.241.52.0	84.241.55.255	Aria Shatel PJSC
84.241.56.0	84.241.63.255	Aria Shatel PJSC
85.9.64.0	85.9.67.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
85.9.68.0	85.9.68.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
85.9.69.0	85.9.69.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
85.9.70.0	85.9.71.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
85.9.72.0	85.9.75.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
85.9.76.0	85.9.76.255	Pishgaman Tejarat Sayar Company (Private Joint Stock)
85.9.77.0	85.9.77.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
85.9.78.0	85.9.79.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
85.9.80.0	85.9.81.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
85.9.82.0	85.9.82.255	Pishgaman Tejarat Sayar Company (Private Joint Stock)
85.9.83.0	85.9.83.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
85.9.84.0	85.9.85.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
85.9.86.0	85.9.87.255	Web Gostaran Bandar Company (PJS)
85.9.88.0	85.9.91.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
85.9.92.0	85.9.93.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
85.9.94.0	85.9.94.255	Gostaresh Ertebat Azin Kia Company PJSC
85.9.95.0	85.9.95.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
85.9.96.0	85.9.103.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
85.9.104.0	85.9.104.255	Taknet Afzar Aftab PJSC
85.9.105.0	85.9.105.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
85.9.106.0	85.9.106.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
85.9.107.0	85.9.107.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
85.9.108.0	85.9.108.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
85.9.109.0	85.9.109.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
85.9.110.0	85.9.110.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
85.9.111.0	85.9.111.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
85.9.112.0	85.9.112.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
85.9.113.0	85.9.113.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
85.9.114.0	85.9.114.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
85.9.115.0	85.9.115.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
85.9.116.0	85.9.116.255	Web Gostaran Bandar Company (PJS)
85.9.117.0	85.9.117.255	Web Gostaran Bandar Company (PJS)
85.9.118.0	85.9.118.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
85.9.119.0	85.9.119.255	Web Gostaran Bandar Company (PJS)
85.9.120.0	85.9.123.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
85.9.124.0	85.9.127.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
85.15.0.0	85.15.0.255	Aria Shatel PJSC
85.15.1.0	85.15.1.255	Aria Shatel PJSC
85.15.2.0	85.15.2.255	Aria Shatel PJSC
85.15.3.0	85.15.3.255	Aria Shatel PJSC
85.15.4.0	85.15.4.255	Aria Shatel PJSC
85.15.5.0	85.15.5.255	Aria Shatel PJSC
85.15.6.0	85.15.6.255	Aria Shatel PJSC
85.15.7.0	85.15.7.255	Aria Shatel PJSC
85.15.8.0	85.15.15.255	Aria Shatel PJSC
85.15.16.0	85.15.31.255	Aria Shatel PJSC
85.15.32.0	85.15.47.255	Aria Shatel PJSC
85.15.48.0	85.15.63.255	Aria Shatel PJSC
85.133.128.0	85.133.128.255	Sepanta Communication Development Co. Ltd
85.133.129.0	85.133.129.255	Sepanta Communication Development Co. Ltd
85.133.130.0	85.133.130.255	Sepanta Communication Development Co. Ltd
85.133.131.0	85.133.131.255	Sepanta Communication Development Co. Ltd
85.133.132.0	85.133.132.255	Sepanta Communication Development Co. Ltd
85.133.133.0	85.133.133.255	Sepanta Communication Development Co. Ltd
85.133.134.0	85.133.134.255	Sepanta Communication Development Co. Ltd
85.133.135.0	85.133.135.255	Sepanta Communication Development Co. Ltd
85.133.136.0	85.133.136.255	Sepanta Communication Development Co. Ltd
85.133.137.0	85.133.137.255	Sepanta Communication Development Co. Ltd
85.133.138.0	85.133.138.255	Sepanta Communication Development Co. Ltd
85.133.139.0	85.133.139.255	Sepanta Communication Development Co. Ltd
85.133.140.0	85.133.140.255	Sefroyek Pardaz Engineering PJSC
85.133.141.0	85.133.141.255	Sepanta Communication Development Co. Ltd
85.133.142.0	85.133.142.255	Sepanta Communication Development Co. Ltd
85.133.143.0	85.133.143.255	Sepanta Communication Development Co. Ltd
85.133.144.0	85.133.144.255	Sepanta Communication Development Co. Ltd
85.133.145.0	85.133.145.255	Sepanta Communication Development Co. Ltd
85.133.146.0	85.133.146.255	Sepanta Communication Development Co. Ltd
85.133.147.0	85.133.147.255	Sepanta Communication Development Co. Ltd
85.133.148.0	85.133.148.255	Sepanta Communication Development Co. Ltd
85.133.149.0	85.133.149.255	Sepanta Communication Development Co. Ltd
85.133.150.0	85.133.150.255	Sepanta Communication Development Co. Ltd
85.133.151.0	85.133.151.255	Sepanta Communication Development Co. Ltd
85.133.152.0	85.133.152.255	Sepanta Communication Development Co. Ltd
85.133.153.0	85.133.153.255	Mizban Pardazesh Nasle Farda PJSC
85.133.154.0	85.133.154.255	Sepanta Communication Development Co. Ltd
85.133.155.0	85.133.155.255	Sepanta Communication Development Co. Ltd
85.133.156.0	85.133.156.255	Sepanta Communication Development Co. Ltd
85.133.157.0	85.133.157.255	Sepanta Communication Development Co. Ltd
85.133.158.0	85.133.158.255	Sepanta Communication Development Co. Ltd
85.133.159.0	85.133.159.255	Sepanta Communication Development Co. Ltd
85.133.160.0	85.133.160.255	Shabakieh Isfahan Co PJSC
85.133.161.0	85.133.161.255	Shabakieh Isfahan Co PJSC
85.133.162.0	85.133.162.255	Shabakieh Isfahan Co PJSC
85.133.163.0	85.133.163.255	Shabakieh Isfahan Co PJSC
85.133.164.0	85.133.164.255	Sepanta Communication Development Co. Ltd
85.133.165.0	85.133.165.255	Sepanta Communication Development Co. Ltd
85.133.166.0	85.133.166.255	Sepanta Communication Development Co. Ltd
85.133.167.0	85.133.167.255	Sepanta Communication Development Co. Ltd
85.133.168.0	85.133.168.255	Sepanta Communication Development Co. Ltd
85.133.169.0	85.133.169.255	Dade Pardaz Kimia Pouyesh PJS.
85.133.170.0	85.133.170.255	Dade Pardaz Kimia Pouyesh PJS.
85.133.171.0	85.133.171.255	Sepanta Communication Development Co. Ltd
85.133.172.0	85.133.172.255	Sepanta Communication Development Co. Ltd
85.133.173.0	85.133.173.255	Sepanta Communication Development Co. Ltd
85.133.174.0	85.133.174.255	Sepanta Communication Development Co. Ltd
85.133.175.0	85.133.175.255	Sepanta Communication Development Co. Ltd
85.133.176.0	85.133.176.255	Sepanta Communication Development Co. Ltd
85.133.177.0	85.133.177.255	Sepanta Communication Development Co. Ltd
85.133.178.0	85.133.178.255	Sepanta Communication Development Co. Ltd
85.133.179.0	85.133.179.255	Sepanta Communication Development Co. Ltd
85.133.180.0	85.133.180.255	Sepanta Communication Development Co. Ltd
85.133.181.0	85.133.181.255	Sepanta Communication Development Co. Ltd
85.133.182.0	85.133.182.255	Sepanta Communication Development Co. Ltd
85.133.183.0	85.133.183.255	Sepanta Communication Development Co. Ltd
85.133.184.0	85.133.184.255	Sepanta Communication Development Co. Ltd
85.133.185.0	85.133.185.255	Sepanta Communication Development Co. Ltd
85.133.186.0	85.133.186.255	Sepanta Communication Development Co. Ltd
85.133.187.0	85.133.187.255	Sepanta Communication Development Co. Ltd
85.133.188.0	85.133.188.255	Sepanta Communication Development Co. Ltd
85.133.189.0	85.133.189.255	Sepanta Communication Development Co. Ltd
85.133.190.0	85.133.190.255	Sepanta Communication Development Co. Ltd
85.133.191.0	85.133.191.255	Sepanta Communication Development Co. Ltd
85.133.192.0	85.133.192.255	Sepanta Communication Development Co. Ltd
85.133.193.0	85.133.193.255	Amir Hosein Maaref
85.133.195.0	85.133.195.255	Pars Abr Toseeh Ertebatat Ltd
85.133.198.0	85.133.198.255	Sigma IT Infrastructures Development Co. (Ltd.)
85.133.200.0	85.133.200.255	Sepanta Communication Development Co. Ltd
85.133.202.0	85.133.202.255	FanAvaran Mihan Mizban PJSC
85.133.205.0	85.133.205.255	Patron Technology Persia Ltd
85.133.206.0	85.133.206.255	DyarWeb Information Technology Ltd
85.133.207.0	85.133.207.255	Abr Ayande Iranian Co. (Private Joint Stock)
85.133.208.0	85.133.208.255	Sepanta Communication Development Co. Ltd
85.133.209.0	85.133.209.255	Sepanta Communication Development Co. Ltd
85.133.210.0	85.133.210.255	Sepanta Communication Development Co. Ltd
85.133.211.0	85.133.211.255	Sepanta Communication Development Co. Ltd
85.133.212.0	85.133.212.255	Sepanta Communication Development Co. Ltd
85.133.213.0	85.133.213.255	Sepanta Communication Development Co. Ltd
85.133.215.0	85.133.215.255	Abr Ayande Iranian Co. (Private Joint Stock)
85.133.217.0	85.133.217.255	Sepanta Communication Development Co. Ltd
85.133.220.0	85.133.220.255	Abr Ayande Iranian Co. (Private Joint Stock)
85.133.221.0	85.133.221.255	Pars Abr Toseeh Ertebatat Ltd
85.133.222.0	85.133.222.255	Sepanta Communication Development Co. Ltd
85.133.223.0	85.133.223.255	Sepanta Communication Development Co. Ltd
85.133.225.0	85.133.225.255	Pars Abr Toseeh Ertebatat Ltd
85.133.226.0	85.133.226.255	Sepanta Communication Development Co. Ltd
85.133.227.0	85.133.227.255	DyarWeb Information Technology Ltd
85.133.229.0	85.133.229.255	Shabakieh Isfahan Co PJSC
85.133.230.0	85.133.230.255	Pars Abr Toseeh Ertebatat Ltd
85.133.231.0	85.133.231.255	Sepanta Communication Development Co. Ltd
85.133.232.0	85.133.232.255	Sepanta Communication Development Co. Ltd
85.133.233.0	85.133.233.255	Sepanta Communication Development Co. Ltd
85.133.235.0	85.133.235.255	Sepanta Communication Development Co. Ltd
85.133.239.0	85.133.239.255	Sefroyek Pardaz Engineering PJSC
85.133.240.0	85.133.240.255	Respina Networks & Beyond PJSC
85.133.242.0	85.133.242.255	Shabakieh Isfahan Co PJSC
85.133.243.0	85.133.243.255	Sepanta Communication Development Co. Ltd
85.133.244.0	85.133.244.255	Sepanta Communication Development Co. Ltd
85.133.245.0	85.133.245.255	Sepanta Communication Development Co. Ltd
85.133.248.0	85.133.248.255	Pars Abr Toseeh Ertebatat Ltd
85.133.249.0	85.133.249.255	Amir Hosein Maaref
85.133.250.0	85.133.250.255	Pars Abr Toseeh Ertebatat Ltd
85.133.251.0	85.133.251.255	Sepanta Communication Development Co. Ltd
85.133.252.0	85.133.252.255	Sepanta Communication Development Co. Ltd
85.133.254.0	85.133.254.255	DyarWeb Information Technology Ltd
85.133.255.0	85.133.255.255	Sepanta Communication Development Co. Ltd
85.159.113.0	85.159.113.255	Behsa Communication & Development Co. (Ltd.)
85.185.0.0	85.185.3.255	Iran Telecommunication Company Pjs
85.185.4.0	85.185.5.255	Iran Telecommunication Company Pjs
85.185.6.0	85.185.7.255	Iran Telecommunication Company Pjs
85.185.8.0	85.185.11.255	Iran Telecommunication Company Pjs
85.185.12.0	85.185.13.255	Iran Telecommunication Company Pjs
85.185.14.0	85.185.15.255	Iran Telecommunication Company Pjs
85.185.16.0	85.185.17.255	Iran Telecommunication Company Pjs
85.185.18.0	85.185.19.255	Iran Telecommunication Company Pjs
85.185.20.0	85.185.23.255	Iran Telecommunication Company Pjs
85.185.24.0	85.185.27.255	Iran Telecommunication Company Pjs
85.185.28.0	85.185.29.255	Iran Telecommunication Company Pjs
85.185.30.0	85.185.30.255	Iran Telecommunication Company Pjs
85.185.31.0	85.185.31.255	Iran Telecommunication Company Pjs
85.185.32.0	85.185.35.255	Iran Telecommunication Company Pjs
85.185.36.0	85.185.39.255	Iran Cell Service and Communication Company
85.185.40.0	85.185.41.255	Iran Telecommunication Company Pjs
85.185.42.0	85.185.42.255	Iran Telecommunication Company Pjs
85.185.43.0	85.185.43.255	Iran Telecommunication Company Pjs
85.185.44.0	85.185.44.255	Iran Telecommunication Company Pjs
85.185.45.0	85.185.45.255	Iran Information Technology Company PJSC
85.185.46.0	85.185.47.255	Iran Telecommunication Company Pjs
85.185.48.0	85.185.63.255	Iran Telecommunication Company Pjs
85.185.64.0	85.185.67.255	Iran Telecommunication Company Pjs
85.185.68.0	85.185.69.255	Iran Telecommunication Company Pjs
85.185.70.0	85.185.70.255	Telecommunication Infrastructure Company
85.185.71.0	85.185.71.255	Iran Telecommunication Company Pjs
85.185.72.0	85.185.75.255	Iran Telecommunication Company Pjs
85.185.76.0	85.185.79.255	Iran Telecommunication Company Pjs
85.185.80.0	85.185.83.255	Iran Telecommunication Company Pjs
85.185.84.0	85.185.87.255	Iran Telecommunication Company Pjs
85.185.88.0	85.185.91.255	Iran Telecommunication Company Pjs
85.185.92.0	85.185.93.255	Iran Telecommunication Company Pjs
85.185.94.0	85.185.94.255	Iran Telecommunication Company Pjs
85.185.95.0	85.185.95.255	Iran Telecommunication Company Pjs
85.185.96.0	85.185.99.255	Iran Telecommunication Company Pjs
85.185.100.0	85.185.103.255	Iran Telecommunication Company Pjs
85.185.104.0	85.185.107.255	Iran Telecommunication Company Pjs
85.185.108.0	85.185.109.255	Iran Telecommunication Company Pjs
85.185.110.0	85.185.110.255	Iran Telecommunication Company Pjs
85.185.111.0	85.185.111.255	Iran Telecommunication Company Pjs
85.185.112.0	85.185.119.255	Iran Telecommunication Company Pjs
85.185.120.0	85.185.127.255	Iran Telecommunication Company Pjs
85.185.128.0	85.185.131.255	Iran Telecommunication Company Pjs
85.185.132.0	85.185.132.255	Iran Telecommunication Company Pjs
85.185.133.0	85.185.133.255	Iran Telecommunication Company Pjs
85.185.134.0	85.185.135.255	Iran Telecommunication Company Pjs
85.185.136.0	85.185.136.255	Iran Telecommunication Company Pjs
85.185.137.0	85.185.137.255	Iran Telecommunication Company Pjs
85.185.138.0	85.185.139.255	Iran Telecommunication Company Pjs
85.185.140.0	85.185.141.255	Iran Telecommunication Company Pjs
85.185.142.0	85.185.143.255	Iran Telecommunication Company Pjs
85.185.144.0	85.185.151.255	Iran Telecommunication Company Pjs
85.185.152.0	85.185.153.255	Iran Telecommunication Company Pjs
85.185.154.0	85.185.154.255	Iran Telecommunication Company Pjs
85.185.155.0	85.185.155.255	Iran Telecommunication Company Pjs
85.185.156.0	85.185.159.255	Iran Telecommunication Company Pjs
85.185.160.0	85.185.163.255	Iran Telecommunication Company Pjs
85.185.164.0	85.185.165.255	Iran Telecommunication Company Pjs
85.185.166.0	85.185.167.255	Iran Telecommunication Company Pjs
85.185.168.0	85.185.171.255	Iran Telecommunication Company Pjs
85.185.172.0	85.185.172.255	Iran Telecommunication Company Pjs
85.185.173.0	85.185.173.255	Iran Telecommunication Company Pjs
85.185.174.0	85.185.175.255	Iran Telecommunication Company Pjs
85.185.176.0	85.185.183.255	Iran Telecommunication Company Pjs
85.185.184.0	85.185.185.255	Iran Telecommunication Company Pjs
85.185.186.0	85.185.186.255	Iran Telecommunication Company Pjs
85.185.187.0	85.185.187.255	Iran Telecommunication Company Pjs
85.185.188.0	85.185.191.255	Iran Telecommunication Company Pjs
85.185.192.0	85.185.199.255	Iran Telecommunication Company Pjs
85.185.200.0	85.185.201.255	Iran Telecommunication Company Pjs
85.185.202.0	85.185.202.255	Iran Telecommunication Company Pjs
85.185.203.0	85.185.203.255	Iran Telecommunication Company Pjs
85.185.204.0	85.185.205.255	Iran Telecommunication Company Pjs
85.185.206.0	85.185.206.255	Iran Telecommunication Company Pjs
85.185.207.0	85.185.207.255	Iran Telecommunication Company Pjs
85.185.208.0	85.185.209.255	Iran Telecommunication Company Pjs
85.185.210.0	85.185.210.255	Iran Telecommunication Company Pjs
85.185.211.0	85.185.211.255	Iran Telecommunication Company Pjs
85.185.212.0	85.185.215.255	Iran Telecommunication Company Pjs
85.185.216.0	85.185.219.255	Iran Telecommunication Company Pjs
85.185.220.0	85.185.220.255	Iran Telecommunication Company Pjs
85.185.221.0	85.185.221.255	Iran Telecommunication Company Pjs
85.185.222.0	85.185.223.255	Iran Telecommunication Company Pjs
85.185.224.0	85.185.227.255	Iran Telecommunication Company Pjs
85.185.228.0	85.185.229.255	Iran Telecommunication Company Pjs
85.185.230.0	85.185.230.255	Iran Telecommunication Company Pjs
85.185.231.0	85.185.231.255	Iran Telecommunication Company Pjs
85.185.232.0	85.185.233.255	Iran Telecommunication Company Pjs
85.185.234.0	85.185.234.255	Iran Telecommunication Company Pjs
85.185.235.0	85.185.235.255	Iran Telecommunication Company Pjs
85.185.236.0	85.185.239.255	Iran Telecommunication Company Pjs
85.185.240.0	85.185.241.255	Iran Telecommunication Company Pjs
85.185.242.0	85.185.243.255	Iran Telecommunication Company Pjs
85.185.244.0	85.185.247.255	Iran Telecommunication Company Pjs
85.185.248.0	85.185.249.255	Iran Telecommunication Company Pjs
85.185.250.0	85.185.251.255	Iran Telecommunication Company Pjs
85.185.252.0	85.185.253.255	Iran Telecommunication Company Pjs
85.185.254.0	85.185.254.255	Iran Telecommunication Company Pjs
85.185.255.0	85.185.255.255	Iran Telecommunication Company Pjs
85.198.0.0	85.198.0.255	Tarashe Sabz Tehran Co. Ltd.
85.198.1.0	85.198.1.255	Tarashe Sabz Tehran Co. Ltd.
85.198.2.0	85.198.2.255	Tarashe Sabz Tehran Co. Ltd.
85.198.3.0	85.198.3.255	Tarashe Sabz Tehran Co. Ltd.
85.198.4.0	85.198.4.255	Tarashe Sabz Tehran Co. Ltd.
85.198.5.0	85.198.5.255	Tarashe Sabz Tehran Co. Ltd.
85.198.6.0	85.198.6.255	Tarashe Sabz Tehran Co. Ltd.
85.198.7.0	85.198.7.255	Tarashe Sabz Tehran Co. Ltd.
85.198.8.0	85.198.11.255	Asre Dadeha Asiatech
85.198.12.0	85.198.15.255	Asre Dadeha Asiatech
85.198.16.0	85.198.17.255	Asre Dadeha Asiatech
85.198.18.0	85.198.18.255	Asiatech Data Transmission Company
85.198.19.0	85.198.19.255	Asre Dadeha Asiatech
85.198.20.0	85.198.21.255	Asre Dadeha Asiatech
85.198.22.0	85.198.23.255	Asre Dadeha Asiatech
85.198.24.0	85.198.24.255	Sefroyek Pardaz Engineering PJSC
85.198.25.0	85.198.25.255	Sefroyek Pardaz Engineering PJSC
85.198.26.0	85.198.26.255	Sefroyek Pardaz Engineering PJSC
85.198.27.0	85.198.27.255	Sefroyek Pardaz Engineering PJSC
85.198.28.0	85.198.29.255	Asiatech Data Transmission Company
85.198.30.0	85.198.31.255	Asiatech Data Transmission Company
85.198.48.0	85.198.63.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
85.204.30.0	85.204.31.255	Parvaresh Dadeha Co. Private Joint Stock
85.204.76.0	85.204.77.255	Parvaresh Dadeha Co. Private Joint Stock
85.204.80.0	85.204.95.255	Iran Telecommunication Company Pjs
85.204.104.0	85.204.105.255	Parvaresh Dadeha Co. Private Joint Stock
85.204.128.0	85.204.131.255	Iran Telecommunication Company Pjs
85.204.208.0	85.204.223.255	Iran Telecommunication Company Pjs
85.208.252.0	85.208.252.255	Green Web Samaneh Novin PJSC
85.208.253.0	85.208.253.255	Green Web Samaneh Novin PJSC
85.208.254.0	85.208.254.255	Green Web Samaneh Novin PJSC
85.208.255.0	85.208.255.255	Green Web Samaneh Novin PJSC
85.209.40.0	85.209.40.255	Rayanesh Pardis Saman PJSC
85.209.41.0	85.209.41.255	Mofid Brokerage PJSC
85.239.192.0	85.239.192.255	Mobile Communication Company of Iran Plc
85.239.193.0	85.239.193.255	Mobile Communication Company of Iran Plc
85.239.194.0	85.239.195.255	Mobile Communication Company of Iran Plc
85.239.196.0	85.239.199.255	Mobile Communication Company of Iran Plc
85.239.200.0	85.239.207.255	Mobile Communication Company of Iran Plc
85.239.208.0	85.239.223.255	Mobile Communication Company of Iran Plc
86.55.0.0	86.55.15.255	Mobile Communication Company of Iran Plc
86.55.16.0	86.55.31.255	Mobile Communication Company of Iran Plc
86.55.32.0	86.55.47.255	Mobile Communication Company of Iran Plc
86.55.48.0	86.55.63.255	Mobile Communication Company of Iran Plc
86.55.64.0	86.55.79.255	Mobile Communication Company of Iran Plc
86.55.80.0	86.55.95.255	Mobile Communication Company of Iran Plc
86.55.96.0	86.55.111.255	Mobile Communication Company of Iran Plc
86.55.112.0	86.55.127.255	Mobile Communication Company of Iran Plc
86.55.128.0	86.55.128.255	Mobile Communication Company of Iran Plc
86.55.129.0	86.55.129.255	Mobile Communication Company of Iran Plc
86.55.130.0	86.55.131.255	Mobile Communication Company of Iran Plc
86.55.132.0	86.55.135.255	Mobile Communication Company of Iran Plc
86.55.136.0	86.55.143.255	Mobile Communication Company of Iran Plc
86.55.144.0	86.55.159.255	Mobile Communication Company of Iran Plc
86.55.160.0	86.55.175.255	Mobile Communication Company of Iran Plc
86.55.176.0	86.55.191.255	Mobile Communication Company of Iran Plc
86.55.192.0	86.55.207.255	Mobile Communication Company of Iran Plc
86.55.208.0	86.55.223.255	Mobile Communication Company of Iran Plc
86.55.224.0	86.55.239.255	Mobile Communication Company of Iran Plc
86.55.240.0	86.55.255.255	Mobile Communication Company of Iran Plc
86.57.0.0	86.57.3.255	Asiatech Data Transmission Company
86.57.4.0	86.57.7.255	Asiatech Data Transmission Company
86.57.8.0	86.57.11.255	Asiatech Data Transmission Company
86.57.12.0	86.57.15.255	Asiatech Data Transmission Company
86.57.16.0	86.57.19.255	Asiatech Data Transmission Company
86.57.20.0	86.57.23.255	Asiatech Data Transmission Company
86.57.24.0	86.57.27.255	Asiatech Data Transmission Company
86.57.28.0	86.57.31.255	Asiatech Data Transmission Company
86.57.32.0	86.57.35.255	Asiatech Data Transmission Company
86.57.36.0	86.57.39.255	Asiatech Data Transmission Company
86.57.40.0	86.57.43.255	Asiatech Data Transmission Company
86.57.44.0	86.57.47.255	Asiatech Data Transmission Company
86.57.48.0	86.57.51.255	Asiatech Data Transmission Company
86.57.52.0	86.57.55.255	Asiatech Data Transmission Company
86.57.56.0	86.57.59.255	Asiatech Data Transmission Company
86.57.60.0	86.57.63.255	Asiatech Data Transmission Company
86.57.64.0	86.57.67.255	Asiatech Data Transmission Company
86.57.68.0	86.57.71.255	Asiatech Data Transmission Company
86.57.72.0	86.57.75.255	Asiatech Data Transmission Company
86.57.76.0	86.57.79.255	Asiatech Data Transmission Company
86.57.80.0	86.57.83.255	Asiatech Data Transmission Company
86.57.84.0	86.57.87.255	Asiatech Data Transmission Company
86.57.88.0	86.57.91.255	Asiatech Data Transmission Company
86.57.92.0	86.57.95.255	Asiatech Data Transmission Company
86.57.96.0	86.57.99.255	Asiatech Data Transmission Company
86.57.100.0	86.57.103.255	Asiatech Data Transmission Company
86.57.104.0	86.57.107.255	Asiatech Data Transmission Company
86.57.108.0	86.57.111.255	Asiatech Data Transmission Company
86.57.112.0	86.57.115.255	Asiatech Data Transmission Company
86.57.116.0	86.57.119.255	Asiatech Data Transmission Company
86.57.120.0	86.57.123.255	Asiatech Data Transmission Company
86.57.124.0	86.57.127.255	Asiatech Data Transmission Company
86.104.32.0	86.104.32.255	Afranet
86.104.33.0	86.104.33.255	Afranet
86.104.34.0	86.104.34.255	Afranet
86.104.35.0	86.104.35.255	Afranet
86.104.36.0	86.104.36.255	Afranet
86.104.37.0	86.104.37.255	Afranet
86.104.38.0	86.104.38.255	Afranet
86.104.39.0	86.104.39.255	Afranet
86.104.40.0	86.104.40.255	Afranet
86.104.41.0	86.104.41.255	Afranet
86.104.42.0	86.104.42.255	Afranet
86.104.43.0	86.104.43.255	Afranet
86.104.44.0	86.104.44.255	Afranet
86.104.45.0	86.104.45.255	Afranet
86.104.46.0	86.104.46.255	Afranet
86.104.47.0	86.104.47.255	Afranet
86.104.80.0	86.104.95.255	Iran Telecommunication Company Pjs
86.104.96.0	86.104.111.255	Iran Telecommunication Company Pjs
86.104.232.0	86.104.239.255	Parvaresh Dadeha Co. Private Joint Stock
86.104.240.0	86.104.247.255	Iran Telecommunication Company Pjs
86.105.40.0	86.105.47.255	Iran Telecommunication Company Pjs
86.105.128.0	86.105.143.255	Iran Telecommunication Company Pjs
86.106.142.0	86.106.142.255	Netmihan Communication Company Ltd
86.106.192.0	86.106.199.255	Iran Telecommunication Company Pjs
86.107.0.0	86.107.0.255	Mobile Communication Company of Iran Plc
86.107.1.0	86.107.1.255	Mobile Communication Company of Iran Plc
86.107.2.0	86.107.3.255	Mobile Communication Company of Iran Plc
86.107.4.0	86.107.7.255	Mobile Communication Company of Iran Plc
86.107.8.0	86.107.15.255	Mobile Communication Company of Iran Plc
86.107.47.0	86.107.47.255	Netmihan Communication Company Ltd
86.107.80.0	86.107.95.255	Iran Telecommunication Company Pjs
86.107.144.0	86.107.159.255	Iran Telecommunication Company Pjs
86.107.172.0	86.107.175.255	Iran Telecommunication Company Pjs
86.107.184.0	86.107.184.255	Netmihan Communication Company Ltd
86.107.208.0	86.107.208.255	Mobile Communication Company of Iran Plc
86.107.209.0	86.107.209.255	Mobile Communication Company of Iran Plc
86.107.210.0	86.107.210.255	Mobile Communication Company of Iran Plc
86.107.211.0	86.107.211.255	Mobile Communication Company of Iran Plc
86.107.212.0	86.107.215.255	Mobile Communication Company of Iran Plc
86.107.216.0	86.107.223.255	Mobile Communication Company of Iran Plc
86.109.32.0	86.109.33.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
86.109.34.0	86.109.34.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
86.109.35.0	86.109.35.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
86.109.36.0	86.109.39.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
86.109.40.0	86.109.40.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
86.109.41.0	86.109.41.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
86.109.42.0	86.109.42.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
86.109.43.0	86.109.43.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
86.109.44.0	86.109.45.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
86.109.46.0	86.109.47.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
86.109.48.0	86.109.51.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
86.109.52.0	86.109.52.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
86.109.53.0	86.109.53.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
86.109.54.0	86.109.55.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
86.109.56.0	86.109.59.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
86.109.60.0	86.109.60.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
86.109.61.0	86.109.61.255	Khatam Academic Institute
86.109.62.0	86.109.63.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
87.107.0.0	87.107.3.255	Soroush Rasanheh Company Ltd
87.107.4.0	87.107.4.255	Soroush Rasanheh Company Ltd
87.107.5.0	87.107.5.255	Dade Pardazi Mobinhost Co Ltd
87.107.6.0	87.107.6.255	Dade Pardazi Mobinhost Co Ltd
87.107.7.0	87.107.7.255	Dade Pardazi Mobinhost Co Ltd
87.107.8.0	87.107.8.255	Dade Pardazi Mobinhost Co Ltd
87.107.9.0	87.107.9.255	Dade Pardazi Mobinhost Co Ltd
87.107.10.0	87.107.10.255	Pars Shabakeh Azarakhsh LLC
87.107.11.0	87.107.11.255	Pars Shabakeh Azarakhsh LLC
87.107.12.0	87.107.12.255	Pars Shabakeh Azarakhsh LLC
87.107.13.0	87.107.13.255	Pars Shabakeh Azarakhsh LLC
87.107.14.0	87.107.14.255	Soroush Rasanheh Company Ltd
87.107.15.0	87.107.15.255	Soroush Rasanheh Company Ltd
87.107.16.0	87.107.16.255	Respina Networks & Beyond PJSC
87.107.17.0	87.107.17.255	Respina Networks & Beyond PJSC
87.107.18.0	87.107.18.255	Respina Networks & Beyond PJSC
87.107.19.0	87.107.19.255	Respina Networks & Beyond PJSC
87.107.20.0	87.107.20.255	Imam Khomeini Airport city company
87.107.21.0	87.107.21.255	Imam Khomeini Airport city company
87.107.22.0	87.107.22.255	Imam Khomeini Airport city company
87.107.23.0	87.107.23.255	Imam Khomeini Airport city company
87.107.24.0	87.107.27.255	Soroush Rasanheh Company Ltd
87.107.28.0	87.107.28.255	Respina Networks & Beyond PJSC
87.107.29.0	87.107.29.255	Respina Networks & Beyond PJSC
87.107.30.0	87.107.30.255	Respina Networks & Beyond PJSC
87.107.31.0	87.107.31.255	Respina Networks & Beyond PJSC
87.107.32.0	87.107.33.255	Respina Networks & Beyond PJSC
87.107.34.0	87.107.35.255	Soroush Rasanheh Company Ltd
87.107.36.0	87.107.36.255	Respina Networks & Beyond PJSC
87.107.37.0	87.107.37.255	Respina Networks & Beyond PJSC
87.107.38.0	87.107.38.255	Respina Networks & Beyond PJSC
87.107.39.0	87.107.39.255	Toesegaran Shabakeh Arseh Novin Ltd
87.107.40.0	87.107.41.255	Soroush Rasanheh Company Ltd
87.107.42.0	87.107.43.255	Soroush Rasanheh Company Ltd
87.107.44.0	87.107.44.255	Respina Networks & Beyond PJSC
87.107.45.0	87.107.45.255	Respina Networks & Beyond PJSC
87.107.46.0	87.107.46.255	Respina Networks & Beyond PJSC
87.107.47.0	87.107.47.255	Respina Networks & Beyond PJSC
87.107.48.0	87.107.49.255	Respina Networks & Beyond PJSC
87.107.50.0	87.107.50.255	Road Maintenance & Transportation Organization
87.107.51.0	87.107.51.255	Soroush Rasanheh Company Ltd
87.107.52.0	87.107.53.255	Soroush Rasanheh Company Ltd
87.107.54.0	87.107.54.255	Pars Shabakeh Azarakhsh LLC
87.107.55.0	87.107.55.255	Pars Shabakeh Azarakhsh LLC
87.107.56.0	87.107.56.255	Shiraz University
87.107.57.0	87.107.57.255	Shiraz University
87.107.58.0	87.107.58.255	Shiraz University
87.107.59.0	87.107.59.255	Shiraz University
87.107.60.0	87.107.60.255	Shiraz University
87.107.61.0	87.107.61.255	Shiraz University
87.107.62.0	87.107.62.255	Shiraz University
87.107.63.0	87.107.63.255	Shiraz University
87.107.64.0	87.107.64.255	Respina Networks & Beyond PJSC
87.107.65.0	87.107.65.255	Respina Networks & Beyond PJSC
87.107.66.0	87.107.67.255	Soroush Rasanheh Company Ltd
87.107.68.0	87.107.71.255	Soroush Rasanheh Company Ltd
87.107.72.0	87.107.72.255	Respina Networks & Beyond PJSC
87.107.73.0	87.107.73.255	Respina Networks & Beyond PJSC
87.107.74.0	87.107.74.255	Respina Networks & Beyond PJSC
87.107.75.0	87.107.75.255	Respina Networks & Beyond PJSC
87.107.76.0	87.107.76.255	Respina Networks & Beyond PJSC
87.107.77.0	87.107.77.255	Respina Networks & Beyond PJSC
87.107.78.0	87.107.78.255	Respina Networks & Beyond PJSC
87.107.79.0	87.107.79.255	Respina Networks & Beyond PJSC
87.107.80.0	87.107.80.255	Respina Networks & Beyond PJSC
87.107.81.0	87.107.81.255	FanAvaran Mihan Mizban PJSC
87.107.82.0	87.107.82.255	FanAvaran Mihan Mizban PJSC
87.107.83.0	87.107.83.255	FanAvaran Mihan Mizban PJSC
87.107.84.0	87.107.84.255	Respina Networks & Beyond PJSC
87.107.85.0	87.107.85.255	Respina Networks & Beyond PJSC
87.107.86.0	87.107.86.255	Respina Networks & Beyond PJSC
87.107.87.0	87.107.87.255	Respina Networks & Beyond PJSC
87.107.88.0	87.107.91.255	Soroush Rasanheh Company Ltd
87.107.92.0	87.107.92.255	Soroush Rasanheh Company Ltd
87.107.93.0	87.107.93.255	Secure Infrastructure of Transactional Services Company PJSC
87.107.94.0	87.107.94.255	Soroush Rasanheh Company Ltd
87.107.95.0	87.107.95.255	Respina Networks & Beyond PJSC
87.107.96.0	87.107.99.255	Soroush Rasanheh Company Ltd
87.107.100.0	87.107.100.255	Asre Pardazeshe Ettelaate Amin Institute
87.107.101.0	87.107.101.255	Parsway Shomal Company Ltd
87.107.102.0	87.107.102.255	Rasaneh Avabarid Private Joint Stock Company
87.107.103.0	87.107.103.255	Respina Networks & Beyond PJSC
87.107.104.0	87.107.104.255	Pars Shabakeh Azarakhsh LLC
87.107.105.0	87.107.105.255	Pars Shabakeh Azarakhsh LLC
87.107.106.0	87.107.107.255	Soroush Rasanheh Company Ltd
87.107.108.0	87.107.108.255	Dade Pardazi Mobinhost Co Ltd
87.107.109.0	87.107.109.255	Dade Pardazi Mobinhost Co Ltd
87.107.110.0	87.107.110.255	Dade Pardazi Mobinhost Co Ltd
87.107.111.0	87.107.111.255	Dade Pardazi Mobinhost Co Ltd
87.107.112.0	87.107.115.255	Soroush Rasanheh Company Ltd
87.107.116.0	87.107.116.255	Institute for Research in Fundamental Sciences
87.107.117.0	87.107.117.255	Institute for Research in Fundamental Sciences
87.107.118.0	87.107.119.255	Soroush Rasanheh Company Ltd
87.107.120.0	87.107.127.255	Soroush Rasanheh Company Ltd
87.107.128.0	87.107.131.255	Soroush Rasanheh Company Ltd
87.107.132.0	87.107.133.255	Soroush Rasanheh Company Ltd
87.107.134.0	87.107.134.255	Pardakht Nowin Arian PJSC
87.107.135.0	87.107.135.255	Pardakht Nowin Arian PJSC
87.107.136.0	87.107.136.255	Respina Networks & Beyond PJSC
87.107.137.0	87.107.137.255	Respina Networks & Beyond PJSC
87.107.138.0	87.107.138.255	Respina Networks & Beyond PJSC
87.107.139.0	87.107.139.255	Respina Networks & Beyond PJSC
87.107.140.0	87.107.140.255	Respina Networks & Beyond PJSC
87.107.141.0	87.107.141.255	Respina Networks & Beyond PJSC
87.107.142.0	87.107.142.255	Respina Networks & Beyond PJSC
87.107.143.0	87.107.143.255	Respina Networks & Beyond PJSC
87.107.144.0	87.107.144.255	Dade Pardazi Mobinhost Co Ltd
87.107.145.0	87.107.145.255	Soroush Rasanheh Company Ltd
87.107.146.0	87.107.146.255	Pars Shabakeh Azarakhsh LLC
87.107.147.0	87.107.147.255	Pars Shabakeh Azarakhsh LLC
87.107.148.0	87.107.148.255	Respina Networks & Beyond PJSC
87.107.149.0	87.107.149.255	Respina Networks & Beyond PJSC
87.107.150.0	87.107.150.255	Respina Networks & Beyond PJSC
87.107.151.0	87.107.151.255	Respina Networks & Beyond PJSC
87.107.152.0	87.107.152.255	Dade Pardazi Mobinhost Co Ltd
87.107.153.0	87.107.153.255	Dade Pardazi Mobinhost Co Ltd
87.107.154.0	87.107.154.255	Pars Shabakeh Azarakhsh LLC
87.107.155.0	87.107.155.255	Pars Shabakeh Azarakhsh LLC
87.107.156.0	87.107.157.255	Respina Networks & Beyond PJSC
87.107.158.0	87.107.158.255	Respina Networks & Beyond PJSC
87.107.159.0	87.107.159.255	Sarmayeh Bank Public Joint Stock
87.107.160.0	87.107.163.255	Soroush Rasanheh Company Ltd
87.107.164.0	87.107.164.255	Pars Shabakeh Azarakhsh LLC
87.107.165.0	87.107.165.255	Pars Shabakeh Azarakhsh LLC
87.107.166.0	87.107.166.255	Dade Pardazi Mobinhost Co Ltd
87.107.167.0	87.107.167.255	Dade Pardazi Mobinhost Co Ltd
87.107.168.0	87.107.171.255	Soroush Rasanheh Company Ltd
87.107.172.0	87.107.172.255	Rahkar Pardazesh Khavarmianeh PJSC
87.107.173.0	87.107.173.255	Soroush Rasanheh Company Ltd
87.107.174.0	87.107.174.255	Dade Pardazi Mobinhost Co Ltd
87.107.175.0	87.107.175.255	Dade Pardazi Mobinhost Co Ltd
87.107.176.0	87.107.176.255	Shiraz University
87.107.177.0	87.107.177.255	Shiraz University
87.107.178.0	87.107.178.255	Shiraz University
87.107.179.0	87.107.179.255	Shiraz University
87.107.180.0	87.107.180.255	Shiraz University
87.107.181.0	87.107.181.255	Shiraz University
87.107.182.0	87.107.182.255	Shiraz University
87.107.183.0	87.107.183.255	Shiraz University
87.107.184.0	87.107.184.255	Respina Networks & Beyond PJSC
87.107.185.0	87.107.185.255	Respina Networks & Beyond PJSC
87.107.186.0	87.107.186.255	University of Sistan and Baluchestan
87.107.187.0	87.107.187.255	Respina Networks & Beyond PJSC
87.107.188.0	87.107.189.255	Soroush Rasanheh Company Ltd
87.107.190.0	87.107.190.255	Mizban Dade Shetaban Co. (Ltd)
87.107.191.0	87.107.191.255	Toosee Ertebatat Damavand
87.107.192.0	87.107.192.255	Giti Secure Cloud LLC
87.107.193.0	87.107.193.255	Giti Secure Cloud LLC
87.107.194.0	87.107.194.255	Giti Secure Cloud LLC
87.107.195.0	87.107.195.255	Giti Secure Cloud LLC
87.107.196.0	87.107.196.255	Giti Secure Cloud LLC
87.107.197.0	87.107.197.255	Giti Secure Cloud LLC
87.107.198.0	87.107.198.255	Giti Secure Cloud LLC
87.107.199.0	87.107.199.255	Soroush Rasanheh Company Ltd
87.107.200.0	87.107.207.255	Soroush Rasanheh Company Ltd
87.107.208.0	87.107.223.255	Soroush Rasanheh Company Ltd
87.107.224.0	87.107.255.255	Kish Cell Pars Co. (PJS)
87.236.38.0	87.236.38.255	Pars Abr Toseeh Ertebatat Ltd
87.236.39.0	87.236.39.255	Aria Shatel PJSC
87.236.208.0	87.236.208.255	FanAvaran Mihan Mizban PJSC
87.236.209.0	87.236.209.255	Dade Pardazi Mobinhost Co Ltd
87.236.210.0	87.236.210.255	Aria Web Development LLC
87.236.211.0	87.236.211.255	Dadeh Gostar Sina Shiraz PJSC
87.236.212.0	87.236.212.255	Dade Pardazi Mobinhost Co Ltd
87.236.213.0	87.236.213.255	Mohammad Sheikhe Sajadie
87.236.214.0	87.236.214.255	Mohammad Sheikhe Sajadie
87.236.215.0	87.236.215.255	Pars Shabakeh Azarakhsh LLC
87.247.172.0	87.247.172.255	Hezardastan Unit Cloud Computing PJSC
87.247.173.0	87.247.173.255	Hezardastan Unit Cloud Computing PJSC
87.247.176.0	87.247.176.255	Moasese Gostaresh Etelaat Va Ertebatat Farhangi Neda Rayaneh
87.247.177.0	87.247.177.255	Moasese Gostaresh Etelaat Va Ertebatat Farhangi Neda Rayaneh
87.247.179.0	87.247.179.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
87.247.180.0	87.247.181.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
87.247.182.0	87.247.183.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
87.248.128.0	87.248.128.255	Tose'Eh Ertebatat Novin Aria Co PJS
87.248.129.0	87.248.129.255	Morteza Imani Bidgoli
87.248.130.0	87.248.130.255	Pars Shabakeh Azarakhsh LLC
87.248.131.0	87.248.131.255	Pars Shabakeh Azarakhsh LLC
87.248.132.0	87.248.132.255	Tose'Eh Ertebatat Novin Aria Co PJS
87.248.133.0	87.248.133.255	Pars Shabakeh Azarakhsh LLC
87.248.137.0	87.248.137.255	Pars Shabakeh Azarakhsh LLC
87.248.138.0	87.248.138.255	Pars Shabakeh Azarakhsh LLC
87.248.139.0	87.248.139.255	Pars Shabakeh Azarakhsh LLC
87.248.140.0	87.248.140.255	Tose'Eh Ertebatat Novin Aria Co PJS
87.248.141.0	87.248.141.255	Ertebatat Sabet Parsian Co. PJS
87.248.143.0	87.248.143.255	Tose'Eh Ertebatat Novin Aria Co PJS
87.248.144.0	87.248.144.255	Tose'Eh Ertebatat Novin Aria Co PJS
87.248.145.0	87.248.145.255	Pars Shabakeh Azarakhsh LLC
87.248.147.0	87.248.147.255	Asre Pardazeshe Ettelaate Amin Institute
87.248.149.0	87.248.149.255	Morteza Imani Bidgoli
87.248.150.0	87.248.150.255	Pars Shabakeh Azarakhsh LLC
87.248.151.0	87.248.151.255	Pars Shabakeh Azarakhsh LLC
87.248.152.0	87.248.152.255	Pars Shabakeh Azarakhsh LLC
87.248.153.0	87.248.153.255	Pars Shabakeh Azarakhsh LLC
87.248.154.0	87.248.154.255	Ettelaat Fanavarn-E Tookan Co Ltd
87.248.155.0	87.248.155.255	Pars Shabakeh Azarakhsh LLC
87.248.156.0	87.248.156.255	Pars Shabakeh Azarakhsh LLC
87.248.159.0	87.248.159.255	Noavaran Shabakeh Sabz Mehregan (Ltd.)
87.251.128.0	87.251.135.255	Iran Telecommunication Company Pjs
87.251.136.0	87.251.139.255	Iran Telecommunication Company Pjs
87.251.140.0	87.251.143.255	Iran Telecommunication Company Pjs
87.251.144.0	87.251.145.255	Iran Telecommunication Company Pjs
87.251.146.0	87.251.146.255	Iran Telecommunication Company Pjs
87.251.147.0	87.251.147.255	Iran Telecommunication Company Pjs
87.251.148.0	87.251.149.255	Iran Telecommunication Company Pjs
87.251.150.0	87.251.151.255	Iran Telecommunication Company Pjs
87.251.152.0	87.251.159.255	Iran Telecommunication Company Pjs
88.135.32.0	88.135.32.255	Shetab Information Processing Co
88.135.33.0	88.135.33.255	Netafraz Iranian Ltd.
88.135.34.0	88.135.35.255	Netafraz Iranian Ltd.
88.135.36.0	88.135.36.255	Sefroyek Pardaz Engineering PJSC
88.135.37.0	88.135.37.255	Sefroyek Pardaz Engineering PJSC
88.135.38.0	88.135.38.255	Sefroyek Pardaz Engineering PJSC
88.135.39.0	88.135.39.255	Sefroyek Pardaz Engineering PJSC
88.135.40.0	88.135.47.255	Asiatech Data Transmission Company
88.135.68.0	88.135.68.255	Dade Pardazan Fahim Shahin Shahr (Ltd)
88.135.72.0	88.135.72.255	Web Dadeh Paydar Co (Ltd)
88.135.75.0	88.135.75.255	Mizban Gostar Dade Alvand Ltd
88.218.16.0	88.218.16.255	Mizban Dadeh Pardis .Ltd
88.218.18.0	88.218.18.255	Mizban Dadeh Pardis .Ltd
88.218.19.0	88.218.19.255	Shahkar Towse'e Tejarat Mana PJSC
89.23.126.0	89.23.126.255	Dadeh Pardazesh Boroumand Kerman co. LLC
89.32.0.0	89.32.31.255	Rightel Communication Service Company PJS
89.32.96.0	89.32.111.255	Rightel Communication Service Company PJS
89.32.196.0	89.32.197.255	Parvaresh Dadeha Co. Private Joint Stock
89.32.248.0	89.32.248.255	Netmihan Communication Company Ltd
89.32.249.0	89.32.249.255	Netmihan Communication Company Ltd
89.32.250.0	89.32.250.255	Netmihan Communication Company Ltd
89.32.251.0	89.32.251.255	Netmihan Communication Company Ltd
89.33.18.0	89.33.19.255	Parvaresh Dadeha Co. Private Joint Stock
89.33.87.0	89.33.87.255	Netmihan Communication Company Ltd
89.33.100.0	89.33.103.255	Iran Telecommunication Company Pjs
89.33.128.0	89.33.128.255	Amn Electronic Kashef Management Co. PJSC
89.33.129.0	89.33.129.255	Parvaresh Dadeha Co. Private Joint Stock
89.33.204.0	89.33.205.255	Parvaresh Dadeha Co. Private Joint Stock
89.33.234.0	89.33.235.255	Parvaresh Dadeha Co. Private Joint Stock
89.33.240.0	89.33.241.255	Parvaresh Dadeha Co. Private Joint Stock
89.34.20.0	89.34.21.255	Parvaresh Dadeha Co. Private Joint Stock
89.34.32.0	89.34.32.255	Rightel Communication Service Company PJS
89.34.33.0	89.34.33.255	Rightel Communication Service Company PJS
89.34.34.0	89.34.35.255	Rightel Communication Service Company PJS
89.34.36.0	89.34.39.255	Rightel Communication Service Company PJS
89.34.40.0	89.34.47.255	Rightel Communication Service Company PJS
89.34.48.0	89.34.63.255	Rightel Communication Service Company PJS
89.34.88.0	89.34.89.255	Parvaresh Dadeha Co. Private Joint Stock
89.34.94.0	89.34.95.255	Parvaresh Dadeha Co. Private Joint Stock
89.34.128.0	89.34.159.255	Rightel Communication Service Company PJS
89.34.168.0	89.34.169.255	Parvaresh Dadeha Co. Private Joint Stock
89.34.176.0	89.34.177.255	Parvaresh Dadeha Co. Private Joint Stock
89.34.200.0	89.34.201.255	Parvaresh Dadeha Co. Private Joint Stock
89.34.248.0	89.34.255.255	Iran Telecommunication Company Pjs
89.35.58.0	89.35.59.255	Parvaresh Dadeha Co. Private Joint Stock
89.35.64.0	89.35.71.255	Iran Telecommunication Company Pjs
89.35.120.0	89.35.123.255	Iran Telecommunication Company Pjs
89.35.132.0	89.35.133.255	Parvaresh Dadeha Co. Private Joint Stock
89.35.156.0	89.35.157.255	Parvaresh Dadeha Co. Private Joint Stock
89.35.172.0	89.35.172.255	Netmihan Communication Company Ltd
89.35.176.0	89.35.177.255	Parvaresh Dadeha Co. Private Joint Stock
89.35.180.0	89.35.183.255	Iran Telecommunication Company Pjs
89.35.194.0	89.35.195.255	Parvaresh Dadeha Co. Private Joint Stock
89.36.16.0	89.36.17.255	Parvaresh Dadeha Co. Private Joint Stock
89.36.48.0	89.36.63.255	Iran Telecommunication Company Pjs
89.36.96.0	89.36.99.255	Iran Telecommunication Company Pjs
89.36.100.0	89.36.103.255	Iran Telecommunication Company Pjs
89.36.104.0	89.36.105.255	Iran Telecommunication Company Pjs
89.36.106.0	89.36.107.255	Iran Telecommunication Company Pjs
89.36.108.0	89.36.111.255	Iran Telecommunication Company Pjs
89.36.176.0	89.36.191.255	Iran Telecommunication Company Pjs
89.36.194.0	89.36.195.255	Parvaresh Dadeha Co. Private Joint Stock
89.36.226.0	89.36.227.255	Parvaresh Dadeha Co. Private Joint Stock
89.36.252.0	89.36.253.255	Parvaresh Dadeha Co. Private Joint Stock
89.37.0.0	89.37.0.255	Mobin Net Communication Company (Private Joint Stock)
89.37.1.0	89.37.1.255	Mobin Net Communication Company (Private Joint Stock)
89.37.2.0	89.37.3.255	Mobin Net Communication Company (Private Joint Stock)
89.37.4.0	89.37.7.255	Mobin Net Communication Company (Private Joint Stock)
89.37.8.0	89.37.11.255	Mobin Net Communication Company (Private Joint Stock)
89.37.12.0	89.37.12.255	Mobin Net Communication Company (Private Joint Stock)
89.37.13.0	89.37.13.255	Mobin Net Communication Company (Private Joint Stock)
89.37.14.0	89.37.14.255	Mobin Net Communication Company (Private Joint Stock)
89.37.15.0	89.37.15.255	Mobin Net Communication Company (Private Joint Stock)
89.37.30.0	89.37.31.255	Parvaresh Dadeha Co. Private Joint Stock
89.37.42.0	89.37.43.255	Parvaresh Dadeha Co. Private Joint Stock
89.37.102.0	89.37.103.255	Parvaresh Dadeha Co. Private Joint Stock
89.37.144.0	89.37.145.255	Iran Telecommunication Company Pjs
89.37.146.0	89.37.147.255	Iran Telecommunication Company Pjs
89.37.148.0	89.37.149.255	Iran Telecommunication Company Pjs
89.37.150.0	89.37.151.255	Iran Telecommunication Company Pjs
89.37.152.0	89.37.155.255	Iran Telecommunication Company Pjs
89.37.168.0	89.37.171.255	Iran Telecommunication Company Pjs
89.37.198.0	89.37.199.255	Parvaresh Dadeha Co. Private Joint Stock
89.37.208.0	89.37.211.255	Iran Telecommunication Company Pjs
89.37.218.0	89.37.219.255	Parvaresh Dadeha Co. Private Joint Stock
89.37.240.0	89.37.247.255	Iran Telecommunication Company Pjs
89.37.248.0	89.37.255.255	Iran Telecommunication Company Pjs
89.38.24.0	89.38.25.255	Parvaresh Dadeha Co. Private Joint Stock
89.38.80.0	89.38.95.255	Iran Telecommunication Company Pjs
89.38.102.0	89.38.103.255	Parvaresh Dadeha Co. Private Joint Stock
89.38.184.0	89.38.191.255	Iran Telecommunication Company Pjs
89.38.192.0	89.38.199.255	Iran Telecommunication Company Pjs
89.38.212.0	89.38.215.255	Darvag Cloud Infrastructure Innovators Co.pjs
89.38.242.0	89.38.243.255	Parvaresh Dadeha Co. Private Joint Stock
89.38.244.0	89.38.247.255	Iran Telecommunication Company Pjs
89.39.8.0	89.39.11.255	Iran Telecommunication Company Pjs
89.39.186.0	89.39.187.255	Parvaresh Dadeha Co. Private Joint Stock
89.39.208.0	89.39.208.255	Netmihan Communication Company Ltd
89.40.35.0	89.40.35.255	Mobin Net Communication Company (Private Joint Stock)
89.40.38.0	89.40.39.255	Parvaresh Dadeha Co. Private Joint Stock
89.40.65.0	89.40.65.255	Netmihan Communication Company Ltd
89.40.78.0	89.40.79.255	Parvaresh Dadeha Co. Private Joint Stock
89.40.90.0	89.40.91.255	Parvaresh Dadeha Co. Private Joint Stock
89.40.106.0	89.40.107.255	Parvaresh Dadeha Co. Private Joint Stock
89.40.110.0	89.40.111.255	Parvaresh Dadeha Co. Private Joint Stock
89.40.128.0	89.40.129.255	Parvaresh Dadeha Co. Private Joint Stock
89.40.152.0	89.40.159.255	Iran Telecommunication Company Pjs
89.40.240.0	89.40.243.255	Iran Telecommunication Company Pjs
89.40.244.0	89.40.244.255	Iran Telecommunication Company Pjs
89.40.245.0	89.40.245.255	Iran Telecommunication Company Pjs
89.40.246.0	89.40.247.255	Iran Telecommunication Company Pjs
89.40.248.0	89.40.255.255	Iran Telecommunication Company Pjs
89.41.8.0	89.41.15.255	Parvaresh Dadeha Co. Private Joint Stock
89.41.16.0	89.41.19.255	Parvaresh Dadeha Co. Private Joint Stock
89.41.20.0	89.41.23.255	Parvaresh Dadeha Co. Private Joint Stock
89.41.32.0	89.41.33.255	Parvaresh Dadeha Co. Private Joint Stock
89.41.40.0	89.41.41.255	Asiatech Data Transmission Company
89.41.42.0	89.41.43.255	Asiatech Data Transmission Company
89.41.58.0	89.41.59.255	Parvaresh Dadeha Co. Private Joint Stock
89.41.184.0	89.41.187.255	Darvag Cloud Infrastructure Innovators Co.pjs
89.41.192.0	89.41.223.255	Rightel Communication Service Company PJS
89.41.240.0	89.41.247.255	Parvaresh Dadeha Co. Private Joint Stock
89.42.32.0	89.42.33.255	Parvaresh Dadeha Co. Private Joint Stock
89.42.44.0	89.42.44.255	Netafraz Iranian Ltd.
89.42.45.0	89.42.45.255	Netafraz Iranian Ltd.
89.42.46.0	89.42.47.255	Netafraz Iranian Ltd.
89.42.56.0	89.42.57.255	Parvaresh Dadeha Co. Private Joint Stock
89.42.68.0	89.42.69.255	Parvaresh Dadeha Co. Private Joint Stock
89.42.96.0	89.42.97.255	Iran Telecommunication Company Pjs
89.42.98.0	89.42.99.255	Iran Telecommunication Company Pjs
89.42.100.0	89.42.103.255	Iran Telecommunication Company Pjs
89.42.136.0	89.42.136.255	Asre Pardazeshe Ettelaate Amin Institute
89.42.137.0	89.42.137.255	Asre Pardazeshe Ettelaate Amin Institute
89.42.138.0	89.42.138.255	Asre Pardazeshe Ettelaate Amin Institute
89.42.139.0	89.42.139.255	Asre Pardazeshe Ettelaate Amin Institute
89.42.150.0	89.42.151.255	Parvaresh Dadeha Co. Private Joint Stock
89.42.184.0	89.42.191.255	Iran Telecommunication Company Pjs
89.42.196.0	89.42.196.255	Asre Pardazeshe Ettelaate Amin Institute
89.42.197.0	89.42.197.255	Cooperative Afra ertebatat-e-sabet-e Rasa Co
89.42.198.0	89.42.198.255	Cooperative Afra ertebatat-e-sabet-e Rasa Co
89.42.199.0	89.42.199.255	Tose'e Shabakeh Azad Ltd
89.42.208.0	89.42.208.255	Netmihan Communication Company Ltd
89.42.209.0	89.42.209.255	Netmihan Communication Company Ltd
89.42.210.0	89.42.210.255	Netmihan Communication Company Ltd
89.42.211.0	89.42.211.255	Netmihan Communication Company Ltd
89.42.228.0	89.42.229.255	Parvaresh Dadeha Co. Private Joint Stock
89.43.0.0	89.43.3.255	Mobin Net Communication Company (Private Joint Stock)
89.43.4.0	89.43.7.255	Mobin Net Communication Company (Private Joint Stock)
89.43.8.0	89.43.9.255	Mobin Net Communication Company (Private Joint Stock)
89.43.10.0	89.43.11.255	Mobin Net Communication Company (Private Joint Stock)
89.43.12.0	89.43.13.255	Mobin Net Communication Company (Private Joint Stock)
89.43.14.0	89.43.14.255	Mobin Net Communication Company (Private Joint Stock)
89.43.15.0	89.43.15.255	Ayandeh Bank
89.43.36.0	89.43.37.255	Parvaresh Dadeha Co. Private Joint Stock
89.43.70.0	89.43.71.255	Parvaresh Dadeha Co. Private Joint Stock
89.43.88.0	89.43.95.255	Parvaresh Dadeha Co. Private Joint Stock
89.43.96.0	89.43.103.255	Parvaresh Dadeha Co. Private Joint Stock
89.43.144.0	89.43.147.255	Iran Telecommunication Company Pjs
89.43.148.0	89.43.151.255	Iran Telecommunication Company Pjs
89.43.182.0	89.43.183.255	Parvaresh Dadeha Co. Private Joint Stock
89.43.188.0	89.43.189.255	Parvaresh Dadeha Co. Private Joint Stock
89.43.204.0	89.43.205.255	Parvaresh Dadeha Co. Private Joint Stock
89.43.216.0	89.43.223.255	Iran Telecommunication Company Pjs
89.43.224.0	89.43.231.255	Iran Telecommunication Company Pjs
89.44.112.0	89.44.112.255	Parvaresh Dadeha Co. Private Joint Stock
89.44.113.0	89.44.113.255	Parvaresh Dadeha Co. Private Joint Stock
89.44.118.0	89.44.118.255	Parvaresh Dadeha Co. Private Joint Stock
89.44.119.0	89.44.119.255	Parvaresh Dadeha Co. Private Joint Stock
89.44.128.0	89.44.135.255	Iran Telecommunication Company Pjs
89.44.146.0	89.44.147.255	Parvaresh Dadeha Co. Private Joint Stock
89.44.176.0	89.44.183.255	Iran Telecommunication Company Pjs
89.44.190.0	89.44.191.255	Parvaresh Dadeha Co. Private Joint Stock
89.44.202.0	89.44.203.255	Parvaresh Dadeha Co. Private Joint Stock
89.44.240.0	89.44.241.255	Padidar Technology PJSC
89.44.242.0	89.44.242.255	Pars Abr Toseeh Ertebatat Ltd
89.44.243.0	89.44.243.255	Padidar Technology PJSC
89.45.48.0	89.45.48.255	Mobile Communication Company of Iran Plc
89.45.49.0	89.45.49.255	Mobile Communication Company of Iran Plc
89.45.50.0	89.45.50.255	Mobile Communication Company of Iran Plc
89.45.51.0	89.45.51.255	Mobile Communication Company of Iran Plc
89.45.52.0	89.45.52.255	Mobile Communication Company of Iran Plc
89.45.53.0	89.45.53.255	Mobile Communication Company of Iran Plc
89.45.54.0	89.45.54.255	Mobile Communication Company of Iran Plc
89.45.55.0	89.45.55.255	Mobile Communication Company of Iran Plc
89.45.56.0	89.45.59.255	Mobile Communication Company of Iran Plc
89.45.60.0	89.45.60.255	Mobile Communication Company of Iran Plc
89.45.61.0	89.45.61.255	Mobile Communication Company of Iran Plc
89.45.62.0	89.45.63.255	Mobile Communication Company of Iran Plc
89.45.68.0	89.45.68.255	Parvaresh Dadeha Co. Private Joint Stock
89.45.69.0	89.45.69.255	Parvaresh Dadeha Co. Private Joint Stock
89.45.80.0	89.45.81.255	Parvaresh Dadeha Co. Private Joint Stock
89.45.89.0	89.45.89.255	Netmihan Communication Company Ltd
89.45.112.0	89.45.119.255	Iran Telecommunication Company Pjs
89.45.126.0	89.45.127.255	Parvaresh Dadeha Co. Private Joint Stock
89.45.152.0	89.45.159.255	Iran Telecommunication Company Pjs
89.45.230.0	89.45.231.255	Parvaresh Dadeha Co. Private Joint Stock
89.46.44.0	89.46.45.255	Parvaresh Dadeha Co. Private Joint Stock
89.46.60.0	89.46.61.255	Parvaresh Dadeha Co. Private Joint Stock
89.46.94.0	89.46.95.255	Parvaresh Dadeha Co. Private Joint Stock
89.46.184.0	89.46.191.255	Iran Telecommunication Company Pjs
89.46.216.0	89.46.216.255	Padidar Technology PJSC
89.46.217.0	89.46.217.255	Padidar Technology PJSC
89.46.218.0	89.46.218.255	Padidar Technology PJSC
89.46.219.0	89.46.219.255	Padidar Technology PJSC
89.47.64.0	89.47.71.255	Rightel Communication Service Company PJS
89.47.72.0	89.47.79.255	Rightel Communication Service Company PJS
89.47.128.0	89.47.159.255	Rightel Communication Service Company PJS
89.47.196.0	89.47.196.255	Cooperative Afra ertebatat-e-sabet-e Rasa Co
89.47.199.0	89.47.199.255	Cooperative Afra ertebatat-e-sabet-e Rasa Co
89.47.203.0	89.47.203.255	Cooperative Afra ertebatat-e-sabet-e Rasa Co
89.144.128.0	89.144.129.255	Andishe Sabz Khazar Co. P.J.S.
89.144.130.0	89.144.130.255	Andishe Sabz Khazar Co. P.J.S.
89.144.131.0	89.144.131.255	Andishe Sabz Khazar Co. P.J.S.
89.144.132.0	89.144.135.255	Andishe Sabz Khazar Co. P.J.S.
89.144.136.0	89.144.143.255	Andishe Sabz Khazar Co. P.J.S.
89.144.144.0	89.144.151.255	Andishe Sabz Khazar Co. P.J.S.
89.144.152.0	89.144.155.255	Andishe Sabz Khazar Co. P.J.S.
89.144.156.0	89.144.156.255	Andishe Sabz Khazar Co. P.J.S.
89.144.157.0	89.144.157.255	Andishe Sabz Khazar Co. P.J.S.
89.144.158.0	89.144.159.255	Andishe Sabz Khazar Co. P.J.S.
89.144.160.0	89.144.175.255	Andishe Sabz Khazar Co. P.J.S.
89.144.176.0	89.144.191.255	Andishe Sabz Khazar Co. P.J.S.
89.165.0.0	89.165.15.255	Parvaresh Dadeha Co. Private Joint Stock
89.165.16.0	89.165.23.255	Parvaresh Dadeha Co. Private Joint Stock
89.165.24.0	89.165.31.255	Parvaresh Dadeha Co. Private Joint Stock
89.165.32.0	89.165.47.255	Parvaresh Dadeha Co. Private Joint Stock
89.165.48.0	89.165.55.255	Parvaresh Dadeha Co. Private Joint Stock
89.165.56.0	89.165.63.255	Parvaresh Dadeha Co. Private Joint Stock
89.165.64.0	89.165.127.255	Parvaresh Dadeha Co. Private Joint Stock
89.196.0.0	89.196.0.255	Mobile Communication Company of Iran Plc
89.196.1.0	89.196.1.255	Mobile Communication Company of Iran Plc
89.196.2.0	89.196.3.255	Mobile Communication Company of Iran Plc
89.196.4.0	89.196.7.255	Mobile Communication Company of Iran Plc
89.196.8.0	89.196.15.255	Mobile Communication Company of Iran Plc
89.196.16.0	89.196.31.255	Mobile Communication Company of Iran Plc
89.196.32.0	89.196.47.255	Mobile Communication Company of Iran Plc
89.196.48.0	89.196.63.255	Mobile Communication Company of Iran Plc
89.196.64.0	89.196.79.255	Mobile Communication Company of Iran Plc
89.196.80.0	89.196.95.255	Mobile Communication Company of Iran Plc
89.196.96.0	89.196.111.255	Mobile Communication Company of Iran Plc
89.196.112.0	89.196.127.255	Mobile Communication Company of Iran Plc
89.196.128.0	89.196.128.255	Mobile Communication Company of Iran Plc
89.196.129.0	89.196.129.255	Mobile Communication Company of Iran Plc
89.196.130.0	89.196.131.255	Mobile Communication Company of Iran Plc
89.196.132.0	89.196.135.255	Mobile Communication Company of Iran Plc
89.196.136.0	89.196.143.255	Mobile Communication Company of Iran Plc
89.196.144.0	89.196.159.255	Mobile Communication Company of Iran Plc
89.196.160.0	89.196.175.255	Mobile Communication Company of Iran Plc
89.196.176.0	89.196.191.255	Mobile Communication Company of Iran Plc
89.196.192.0	89.196.207.255	Mobile Communication Company of Iran Plc
89.196.208.0	89.196.223.255	Mobile Communication Company of Iran Plc
89.196.224.0	89.196.239.255	Mobile Communication Company of Iran Plc
89.196.240.0	89.196.255.255	Mobile Communication Company of Iran Plc
89.198.0.0	89.198.0.255	Mobile Communication Company of Iran Plc
89.198.1.0	89.198.1.255	Mobile Communication Company of Iran Plc
89.198.2.0	89.198.3.255	Mobile Communication Company of Iran Plc
89.198.4.0	89.198.7.255	Mobile Communication Company of Iran Plc
89.198.8.0	89.198.15.255	Mobile Communication Company of Iran Plc
89.198.16.0	89.198.31.255	Mobile Communication Company of Iran Plc
89.198.32.0	89.198.47.255	Mobile Communication Company of Iran Plc
89.198.48.0	89.198.63.255	Mobile Communication Company of Iran Plc
89.198.64.0	89.198.79.255	Mobile Communication Company of Iran Plc
89.198.80.0	89.198.95.255	Mobile Communication Company of Iran Plc
89.198.96.0	89.198.111.255	Mobile Communication Company of Iran Plc
89.198.112.0	89.198.127.255	Mobile Communication Company of Iran Plc
89.198.128.0	89.198.143.255	Mobile Communication Company of Iran Plc
89.198.144.0	89.198.159.255	Mobile Communication Company of Iran Plc
89.198.160.0	89.198.175.255	Mobile Communication Company of Iran Plc
89.198.176.0	89.198.191.255	Mobile Communication Company of Iran Plc
89.198.192.0	89.198.207.255	Mobile Communication Company of Iran Plc
89.198.208.0	89.198.223.255	Mobile Communication Company of Iran Plc
89.198.224.0	89.198.239.255	Mobile Communication Company of Iran Plc
89.198.240.0	89.198.255.255	Mobile Communication Company of Iran Plc
89.199.0.0	89.199.15.255	Mobile Communication Company of Iran Plc
89.199.16.0	89.199.31.255	Mobile Communication Company of Iran Plc
89.199.32.0	89.199.47.255	Mobile Communication Company of Iran Plc
89.199.48.0	89.199.63.255	Mobile Communication Company of Iran Plc
89.199.64.0	89.199.79.255	Mobile Communication Company of Iran Plc
89.199.80.0	89.199.95.255	Mobile Communication Company of Iran Plc
89.199.96.0	89.199.111.255	Mobile Communication Company of Iran Plc
89.199.112.0	89.199.127.255	Mobile Communication Company of Iran Plc
89.199.128.0	89.199.128.255	Mobile Communication Company of Iran Plc
89.199.129.0	89.199.129.255	Mobile Communication Company of Iran Plc
89.199.130.0	89.199.131.255	Mobile Communication Company of Iran Plc
89.199.132.0	89.199.135.255	Mobile Communication Company of Iran Plc
89.199.136.0	89.199.143.255	Mobile Communication Company of Iran Plc
89.199.144.0	89.199.159.255	Mobile Communication Company of Iran Plc
89.199.160.0	89.199.175.255	Mobile Communication Company of Iran Plc
89.199.176.0	89.199.191.255	Mobile Communication Company of Iran Plc
89.199.192.0	89.199.207.255	Mobile Communication Company of Iran Plc
89.199.208.0	89.199.223.255	Mobile Communication Company of Iran Plc
89.199.224.0	89.199.239.255	Mobile Communication Company of Iran Plc
89.199.240.0	89.199.255.255	Mobile Communication Company of Iran Plc
89.219.64.0	89.219.127.255	Iran Telecommunication Company Pjs
89.219.192.0	89.219.255.255	Iran Telecommunication Company Pjs
89.221.80.0	89.221.80.255	Ertebat Gostaran Bina PJSC
89.221.81.0	89.221.81.255	Ertebat Gostaran Bina PJSC
89.221.82.0	89.221.82.255	Ertebat Gostaran Bina PJSC
89.221.83.0	89.221.83.255	Ertebat Gostaran Bina PJSC
89.221.84.0	89.221.87.255	Fanava Group
89.221.88.0	89.221.91.255	Dade Samane Fanava Company (PJS)
89.221.92.0	89.221.93.255	Dade Samane Fanava Company (PJS)
89.221.94.0	89.221.94.255	Dade Samane Fanava Company (PJS)
89.221.95.0	89.221.95.255	Fanava Group
89.235.64.0	89.235.64.255	Sadad Informatics Corporation PJSC
89.235.65.0	89.235.65.255	Sadad Informatics Corporation PJSC
89.235.66.0	89.235.66.255	Sadad Informatics Corporation PJSC
89.235.67.0	89.235.67.255	Sadad Informatics Corporation PJSC
89.235.68.0	89.235.68.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
89.235.69.0	89.235.69.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
89.235.70.0	89.235.70.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
89.235.71.0	89.235.71.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
89.235.72.0	89.235.75.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
89.235.76.0	89.235.76.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
89.235.77.0	89.235.77.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
89.235.78.0	89.235.78.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
89.235.79.0	89.235.79.255	Rasaneh Avabarid Private Joint Stock Company
89.235.80.0	89.235.87.255	Iran Telecommunication Company Pjs
89.235.88.0	89.235.95.255	Iran Telecommunication Company Pjs
89.235.96.0	89.235.99.255	Asiatech Data Transmission Company
89.235.100.0	89.235.101.255	Iran Telecommunication Company Pjs
89.235.102.0	89.235.103.255	Iran Telecommunication Company Pjs
89.235.104.0	89.235.107.255	Iran Telecommunication Company Pjs
89.235.108.0	89.235.111.255	Iran Telecommunication Company Pjs
89.235.112.0	89.235.127.255	Iran Telecommunication Company Pjs
91.92.104.0	91.92.104.255	Iran Telecommunication Company Pjs
91.92.114.0	91.92.114.255	Iran Telecommunication Company Pjs
91.92.121.0	91.92.121.255	Iran Telecommunication Company Pjs
91.92.122.0	91.92.123.255	Iran Telecommunication Company Pjs
91.92.124.0	91.92.127.255	Iran Telecommunication Company Pjs
91.92.129.0	91.92.129.255	Iran Telecommunication Company Pjs
91.92.130.0	91.92.131.255	Iran Telecommunication Company Pjs
91.92.132.0	91.92.133.255	Iran Telecommunication Company Pjs
91.92.134.0	91.92.135.255	Iran Telecommunication Company Pjs
91.92.145.0	91.92.145.255	Iran Telecommunication Company Pjs
91.92.146.0	91.92.147.255	Iran Telecommunication Company Pjs
91.92.148.0	91.92.151.255	Iran Telecommunication Company Pjs
91.92.156.0	91.92.159.255	Iran Telecommunication Company Pjs
91.92.164.0	91.92.167.255	Iran Telecommunication Company Pjs
91.92.172.0	91.92.175.255	Iran Telecommunication Company Pjs
91.92.180.0	91.92.183.255	Iran Telecommunication Company Pjs
91.92.184.0	91.92.191.255	Iran Telecommunication Company Pjs
91.92.192.0	91.92.193.255	Iran Telecommunication Company Pjs
91.92.204.0	91.92.207.255	Iran Telecommunication Company Pjs
91.92.208.0	91.92.215.255	Iran Telecommunication Company Pjs
91.92.220.0	91.92.223.255	Iran Telecommunication Company Pjs
91.92.228.0	91.92.229.255	Iran Telecommunication Company Pjs
91.92.231.0	91.92.231.255	Iran Telecommunication Company Pjs
91.92.236.0	91.92.239.255	Iran Telecommunication Company Pjs
91.102.126.0	91.102.126.255	Andisheh Yavaran Tamadon Emrooz LLC
91.102.127.0	91.102.127.255	Varesh Cloud Hosting Limited Liability Company
91.106.64.0	91.106.64.255	PJSC Badr Rayan Jonoob
91.106.65.0	91.106.65.255	PJSC Badr Rayan Jonoob
91.106.66.0	91.106.66.255	PJSC Badr Rayan Jonoob
91.106.67.0	91.106.67.255	PJSC Badr Rayan Jonoob
91.106.68.0	91.106.68.255	PJSC Badr Rayan Jonoob
91.106.69.0	91.106.69.255	PJSC Badr Rayan Jonoob
91.106.70.0	91.106.70.255	PJSC Badr Rayan Jonoob
91.106.71.0	91.106.71.255	PJSC Badr Rayan Jonoob
91.106.72.0	91.106.72.255	PJSC Badr Rayan Jonoob
91.106.73.0	91.106.73.255	PJSC Badr Rayan Jonoob
91.106.74.0	91.106.74.255	PJSC Badr Rayan Jonoob
91.106.75.0	91.106.75.255	PJSC Badr Rayan Jonoob
91.106.76.0	91.106.76.255	PJSC Badr Rayan Jonoob
91.106.77.0	91.106.77.255	PJSC Badr Rayan Jonoob
91.106.78.0	91.106.78.255	PJSC Badr Rayan Jonoob
91.106.79.0	91.106.79.255	PJSC Badr Rayan Jonoob
91.106.80.0	91.106.80.255	PJSC Badr Rayan Jonoob
91.106.81.0	91.106.81.255	PJSC Badr Rayan Jonoob
91.106.82.0	91.106.82.255	PJSC Badr Rayan Jonoob
91.106.83.0	91.106.83.255	PJSC Badr Rayan Jonoob
91.106.84.0	91.106.84.255	PJSC Badr Rayan Jonoob
91.106.85.0	91.106.85.255	PJSC Badr Rayan Jonoob
91.106.86.0	91.106.86.255	PJSC Badr Rayan Jonoob
91.106.87.0	91.106.87.255	PJSC Badr Rayan Jonoob
91.106.88.0	91.106.88.255	PJSC Badr Rayan Jonoob
91.106.89.0	91.106.89.255	PJSC Badr Rayan Jonoob
91.106.90.0	91.106.90.255	PJSC Badr Rayan Jonoob
91.106.91.0	91.106.91.255	PJSC Badr Rayan Jonoob
91.106.92.0	91.106.92.255	PJSC Badr Rayan Jonoob
91.106.93.0	91.106.93.255	PJSC Badr Rayan Jonoob
91.106.94.0	91.106.94.255	PJSC Badr Rayan Jonoob
91.106.95.0	91.106.95.255	PJSC Badr Rayan Jonoob
91.108.128.0	91.108.128.255	Rayaneh Gostar Farzanegan Ahwaz Company LTD.
91.108.129.0	91.108.129.255	Rayaneh Gostar Farzanegan Ahwaz Company LTD.
91.108.130.0	91.108.130.255	Rayaneh Gostar Farzanegan Ahwaz Company LTD.
91.108.131.0	91.108.131.255	Rayaneh Gostar Farzanegan Ahwaz Company LTD.
91.108.132.0	91.108.132.255	Rayaneh Gostar Farzanegan Ahwaz Company LTD.
91.108.133.0	91.108.133.255	Rayaneh Gostar Farzanegan Ahwaz Company LTD.
91.108.134.0	91.108.134.255	Rayaneh Gostar Farzanegan Ahwaz Company LTD.
91.108.135.0	91.108.135.255	Rayaneh Gostar Farzanegan Ahwaz Company LTD.
91.108.136.0	91.108.136.255	Rayaneh Gostar Farzanegan Ahwaz Company LTD.
91.108.137.0	91.108.137.255	Rayaneh Gostar Farzanegan Ahwaz Company LTD.
91.108.138.0	91.108.138.255	Rayaneh Gostar Farzanegan Ahwaz Company LTD.
91.108.139.0	91.108.139.255	Rayaneh Gostar Farzanegan Ahwaz Company LTD.
91.108.140.0	91.108.140.255	Amin Asia Cloud Data PJSC
91.108.141.0	91.108.141.255	Amin Asia Cloud Data PJSC
91.108.142.0	91.108.142.255	Amin Asia Cloud Data PJSC
91.108.143.0	91.108.143.255	Rayaneh Gostar Farzanegan Ahwaz Company LTD.
91.108.144.0	91.108.144.255	Rayaneh Gostar Farzanegan Ahwaz Company LTD.
91.108.145.0	91.108.145.255	Atrin Information & Communications Technology Company PJS
91.108.146.0	91.108.146.255	Atrin Information & Communications Technology Company PJS
91.108.147.0	91.108.147.255	Rayaneh Gostar Farzanegan Ahwaz Company LTD.
91.108.148.0	91.108.148.255	Rayaneh Gostar Farzanegan Ahwaz Company LTD.
91.108.149.0	91.108.149.255	Rayaneh Gostar Farzanegan Ahwaz Company LTD.
91.108.150.0	91.108.150.255	Rayaneh Gostar Farzanegan Ahwaz Company LTD.
91.108.151.0	91.108.151.255	Atrin Information & Communications Technology Company PJS
91.108.152.0	91.108.152.255	Atrin Information & Communications Technology Company PJS
91.108.153.0	91.108.153.255	Rayaneh Gostar Farzanegan Ahwaz Company LTD.
91.108.154.0	91.108.154.255	Rayaneh Gostar Farzanegan Ahwaz Company LTD.
91.108.155.0	91.108.155.255	Rayaneh Gostar Farzanegan Ahwaz Company LTD.
91.108.156.0	91.108.156.255	Rayaneh Gostar Farzanegan Ahwaz Company LTD.
91.108.157.0	91.108.157.255	Rayaneh Gostar Farzanegan Ahwaz Company LTD.
91.108.158.0	91.108.159.255	Rayaneh Gostar Farzanegan Ahwaz Company LTD.
91.109.104.0	91.109.111.255	Iran Telecommunication Company Pjs
91.133.128.0	91.133.143.255	Mobile Communication Company of Iran Plc
91.133.144.0	91.133.159.255	Mobile Communication Company of Iran Plc
91.133.160.0	91.133.175.255	Mobile Communication Company of Iran Plc
91.133.176.0	91.133.191.255	Mobile Communication Company of Iran Plc
91.133.192.0	91.133.207.255	Mobile Communication Company of Iran Plc
91.133.208.0	91.133.223.255	Mobile Communication Company of Iran Plc
91.133.224.0	91.133.239.255	Mobile Communication Company of Iran Plc
91.133.240.0	91.133.255.255	Mobile Communication Company of Iran Plc
91.147.64.0	91.147.79.255	Iran Telecommunication Company Pjs
91.184.64.0	91.184.79.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
91.184.80.0	91.184.95.255	Datak Company LLC
91.185.128.0	91.185.159.255	Iran Telecommunication Company Pjs
91.186.193.0	91.186.193.255	Ettelaat Fanavarn-E Tookan Co Ltd
91.190.88.0	91.190.95.255	Iran Telecommunication Company Pjs
91.192.160.0	91.192.160.255	Maya Pardazesh Iranian LLC
91.194.6.0	91.194.6.255	Majid Soltanifard Razlighi
91.197.242.0	91.197.242.255	Mobtaker Tejarat Novin Vira Ltd
91.198.110.0	91.198.110.255	Respina Networks & Beyond PJSC
91.199.9.0	91.199.9.255	Imen Sanat Novin Alighapoo Private Joint Stock Company
91.199.18.0	91.199.18.255	Imen Sanat Novin Alighapoo Private Joint Stock Company
91.199.27.0	91.199.27.255	Imen Sanat Novin Alighapoo Private Joint Stock Company
91.199.30.0	91.199.30.255	Imen Sanat Novin Alighapoo Private Joint Stock Company
91.199.43.0	91.199.43.255	Web Dadeh Paydar Co (Ltd)
91.199.215.0	91.199.215.255	Parsway Shomal Company Ltd
91.206.171.0	91.206.171.255	Web Dadeh Paydar Co (Ltd)
91.206.177.0	91.206.177.255	Saba Abr Mizban LLC
91.207.18.0	91.207.18.255	Web Dadeh Paydar Co (Ltd)
91.207.138.0	91.207.138.255	Islamic Republic of Iran Post PJSC
91.207.139.0	91.207.139.255	Islamic Republic of Iran Post PJSC
91.207.205.0	91.207.205.255	Asiatech Data Transmission Company
91.208.163.0	91.208.163.255	Amnpardaz Soft Corporation
91.208.165.0	91.208.165.255	Bank Pasargad PJSC
91.209.96.0	91.209.96.255	Ertebat Gostar Amin Koosha Co.ltd.
91.209.161.0	91.209.161.255	Negin Narmafzar Asak Research and Information Cooperative Company
91.209.183.0	91.209.183.255	Mobin Net Communication Company (Private Joint Stock)
91.209.184.0	91.209.184.255	Mobin Net Communication Company (Private Joint Stock)
91.209.186.0	91.209.186.255	Mobin Net Communication Company (Private Joint Stock)
91.209.242.0	91.209.242.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
91.212.16.0	91.212.16.255	Iran Aseman Airlines PJSC
91.212.174.0	91.212.174.255	Vandad Vira Hooman LLC
91.212.232.0	91.212.232.255	Web Dadeh Paydar Co (Ltd)
91.212.252.0	91.212.252.255	Bank Refah Kargaran
91.213.83.0	91.213.83.255	Enteghal Dadeh Mahan Co. PJSC
91.213.151.0	91.213.151.255	Taknet Afzar Aftab PJSC
91.213.157.0	91.213.157.255	Taknet Afzar Aftab PJSC
91.213.164.0	91.213.164.255	Rayan Bours PJSC
91.213.167.0	91.213.167.255	Hesabgar Pardaz Gharb PJSC
91.213.172.0	91.213.172.255	Hesabgar Pardaz Gharb PJSC
91.216.4.0	91.216.4.255	Respina Networks & Beyond PJSC
91.216.71.0	91.216.71.255	Cooperative Afra ertebatat-e-sabet-e Rasa Co
91.216.159.0	91.216.159.255	Mizbani Hooshmand Mehr Afarin PJSC
91.216.171.0	91.216.171.255	Dorsa Expert System PJS
91.216.217.0	91.216.217.255	Web Dadeh Paydar Co (Ltd)
91.217.64.0	91.217.64.255	Sina Bank
91.217.65.0	91.217.65.255	Sina Bank
91.217.177.0	91.217.177.255	Sindad Network Technology PJSC
91.220.79.0	91.220.79.255	MAPNA Power Plants Development & Construction PLC
91.220.113.0	91.220.113.255	Tida Data Processing Ltd
91.220.243.0	91.220.243.255	Ava Katan Qeshm Co PJSC
91.221.116.0	91.221.116.255	Abr Tose'eh Darya Group Company PJSC
91.221.240.0	91.221.240.255	Eligasht Air Travel & Tourist Services Company Limited
91.221.241.0	91.221.241.255	Eligasht Air Travel & Tourist Services Company Limited
91.222.196.0	91.222.196.255	Web Gostaran Bandar Company (PJS)
91.222.197.0	91.222.197.255	Web Gostaran Bandar Company (PJS)
91.222.198.0	91.222.198.255	Nazhin Sepahan IT and Data Processing Plc
91.222.199.0	91.222.199.255	Computer research center of Islamic seminary of Isfahan
91.222.204.0	91.222.207.255	Iran Telecommunication Company Pjs
91.223.61.0	91.223.61.255	Web Dadeh Paydar Co (Ltd)
91.223.116.0	91.223.116.255	Dadeh Gostar Parmis PJS Company
91.223.146.0	91.223.146.255	Parvaz System Information Technology Company (LTD)
91.223.187.0	91.223.187.255	Web Dadeh Paydar Co (Ltd)
91.224.20.0	91.224.21.255	Eligasht Air Travel & Tourist Services Company Limited
91.224.110.0	91.224.110.255	Imam Khomeini Relief Foundation IKRF
91.224.111.0	91.224.111.255	Imam Khomeini Relief Foundation IKRF
91.224.176.0	91.224.177.255	Information Technology Organization of Isfahan Municipality
91.225.52.0	91.225.52.255	IRIB (Islamic Republic of Iran Broadcasting)
91.225.53.0	91.225.53.255	Moavenate Rasaneh Majazi Seda Va Sima
91.225.54.0	91.225.54.255	Moavenate Rasaneh Majazi Seda Va Sima
91.225.55.0	91.225.55.255	Moavenate Rasaneh Majazi Seda Va Sima
91.226.224.0	91.226.225.255	Netafraz Iranian Ltd.
91.226.244.0	91.226.244.255	Abr Gostar Arianet LLC
91.226.246.0	91.226.246.255	Gostaresh Pardazesh Dana Negar Co.(PJSC)
91.227.27.0	91.227.27.255	Enteghal Dadeh Mahan Co. PJSC
91.227.84.0	91.227.85.255	Iran Telecommunication Company Pjs
91.227.86.0	91.227.87.255	Iran Telecommunication Company Pjs
91.227.246.0	91.227.247.255	Ertebatat Fara Gostar Shargh PJSC
91.228.22.0	91.228.23.255	Faragostar Shargh Communication Company PJS
91.228.132.0	91.228.132.255	Fannavaran-e Idea Pardaz-e Saba PJSC
91.228.133.0	91.228.133.255	Fannavaran-e Idea Pardaz-e Saba PJSC
91.228.168.0	91.228.168.255	Pishgam Rayan Dadeh Pajooh PJSC
91.228.189.0	91.228.189.255	Respina Networks & Beyond PJSC
91.228.192.0	91.228.192.255	Saba Abr Mizban LLC
91.229.46.0	91.229.46.255	Fannavaran-e Idea Pardaz-e Saba PJSC
91.229.47.0	91.229.47.255	Fannavaran-e Idea Pardaz-e Saba PJSC
91.229.214.0	91.229.215.255	Rightel Communication Service Company PJS
91.230.32.0	91.230.32.255	Asmanfaraz Sepahan Company (P.J.S)
91.231.144.0	91.231.144.255	Maya Pardazesh Iranian LLC
91.232.64.0	91.232.64.255	Asan Pardakht Persian Co. Ltd
91.232.65.0	91.232.65.255	Asan Pardakht Persian Co. Ltd
91.232.66.0	91.232.66.255	Asan Pardakht Persian Co. Ltd
91.232.67.0	91.232.67.255	Asan Pardakht Persian Co. Ltd
91.232.68.0	91.232.68.255	Asan Pardakht Persian Co. Ltd
91.232.69.0	91.232.69.255	Asan Pardakht Persian Co. Ltd
91.232.72.0	91.232.72.255	Poshtkar Rayaneh Kharg Company PJS
91.232.73.0	91.232.73.255	Parsan Lin Co. PJS
91.232.74.0	91.232.74.255	Poshtkar Rayaneh Kharg Company PJS
91.232.75.0	91.232.75.255	Poshtkar Rayaneh Kharg Company PJS
91.233.56.0	91.233.56.255	Iran Power Generation transmission & Distribution Management Co.
91.233.57.0	91.233.57.255	Iran Power Generation transmission & Distribution Management Co.
91.233.58.0	91.233.58.255	Iran Power Generation transmission & Distribution Management Co.
91.233.59.0	91.233.59.255	Iran Power Generation transmission & Distribution Management Co.
91.234.38.0	91.234.38.255	Web Dadeh Paydar Co (Ltd)
91.234.39.0	91.234.39.255	Vandad Vira Hooman LLC
91.234.52.0	91.234.52.255	Rasaneh Avabarid Private Joint Stock Company
91.234.147.0	91.234.147.255	Khadamat Ertebati Pishkhan Arvin Chenaran Co.llc
91.236.168.0	91.236.168.255	Sefroyek Pardaz Engineering PJSC
91.236.169.0	91.236.169.255	Sefroyek Pardaz Engineering PJSC
91.237.255.0	91.237.255.255	Cooperative Afra ertebatat-e-sabet-e Rasa Co
91.238.0.0	91.238.0.255	Cooperative Afra ertebatat-e-sabet-e Rasa Co
91.239.14.0	91.239.14.255	Trade Promotion Organization of Iran
91.239.59.0	91.239.59.255	Mobin Net Communication Company (Private Joint Stock)
91.239.108.0	91.239.111.255	Productive Distributive Farzanegan Dadeh Pardaz Taha Cooperative Company
91.239.210.0	91.239.210.255	Khalij Fars Ettela Resan LTD
91.239.214.0	91.239.214.255	Taknet Afzar Aftab PJSC
91.240.60.0	91.240.60.255	Nazhin Sepahan IT and Data Processing Plc
91.240.61.0	91.240.61.255	Nazhin Sepahan IT and Data Processing Plc
91.240.62.0	91.240.62.255	Nazhin Sepahan IT and Data Processing Plc
91.240.63.0	91.240.63.255	Nazhin Sepahan IT and Data Processing Plc
91.240.95.0	91.240.95.255	Web Dadeh Paydar Co (Ltd)
91.240.180.0	91.240.180.255	Saman Electronic Payment Kish PJS
91.240.181.0	91.240.181.255	Saman Electronic Payment Kish PJS
91.240.182.0	91.240.182.255	Saman Electronic Payment Kish PJS
91.240.183.0	91.240.183.255	Saman Electronic Payment Kish PJS
91.241.20.0	91.241.21.255	GOSTARESH-E-ERTEBATAT-E MABNA COMPANY (Private Joint Stock)
91.241.92.0	91.241.92.255	Fanavaran Eniac Rayaneh PJSC
91.242.44.0	91.242.45.255	Gostaresh Ertebat Azin Kia Company PJSC
91.243.114.0	91.243.114.255	Shabakeh Ertebatat Artak Towseeh PJSC
91.243.126.0	91.243.127.255	Jahrom Univesity of Medical Sciences
91.243.160.0	91.243.175.255	Sari System Bandarabas Company LLC
91.244.120.0	91.244.123.255	Iran Telecommunication Company Pjs
91.245.228.0	91.245.231.255	Kish Cell Pars Co. (PJS)
91.246.31.0	91.246.31.255	Rasaneh Avabarid Private Joint Stock Company
91.246.44.0	91.246.44.255	Saba Abr Mizban LLC
91.247.66.0	91.247.67.255	Dadeh Pardazan Mojtame Hormoz Company Ltd.
91.247.171.0	91.247.171.255	Sefroyek Pardaz Engineering PJSC
91.250.224.0	91.250.231.255	Iran Telecommunication Company Pjs
91.250.232.0	91.250.239.255	Iran Telecommunication Company Pjs
91.251.0.0	91.251.15.255	Mobile Communication Company of Iran Plc
91.251.16.0	91.251.31.255	Mobile Communication Company of Iran Plc
91.251.32.0	91.251.47.255	Mobile Communication Company of Iran Plc
91.251.48.0	91.251.63.255	Mobile Communication Company of Iran Plc
91.251.64.0	91.251.79.255	Mobile Communication Company of Iran Plc
91.251.80.0	91.251.95.255	Mobile Communication Company of Iran Plc
91.251.96.0	91.251.111.255	Mobile Communication Company of Iran Plc
91.251.112.0	91.251.127.255	Mobile Communication Company of Iran Plc
91.251.128.0	91.251.143.255	Mobile Communication Company of Iran Plc
91.251.144.0	91.251.159.255	Mobile Communication Company of Iran Plc
91.251.160.0	91.251.175.255	Mobile Communication Company of Iran Plc
91.251.176.0	91.251.191.255	Mobile Communication Company of Iran Plc
91.251.192.0	91.251.207.255	Mobile Communication Company of Iran Plc
91.251.208.0	91.251.223.255	Mobile Communication Company of Iran Plc
91.251.224.0	91.251.239.255	Mobile Communication Company of Iran Plc
91.251.240.0	91.251.255.255	Mobile Communication Company of Iran Plc
92.42.48.0	92.42.51.255	Iran Cell Service and Communication Company
92.42.52.0	92.42.55.255	Iran Cell Service and Communication Company
92.42.203.0	92.42.203.255	Respina Networks & Beyond PJSC
92.42.205.0	92.42.205.255	Fanavaran Nicsepehr Zenderood Ltd.
92.42.207.0	92.42.207.255	Giti Secure Cloud LLC
92.43.160.0	92.43.163.255	Iran Telecommunication Company Pjs
92.61.176.0	92.61.176.255	Ministry of Health Treatment and Medical Education
92.61.177.0	92.61.177.255	Ministry of Health Treatment and Medical Education
92.61.178.0	92.61.178.255	Ministry of Health Treatment and Medical Education
92.61.179.0	92.61.179.255	Ministry of Health Treatment and Medical Education
92.61.180.0	92.61.183.255	Abramad Technological Infrastructures Development Company PJS
92.61.184.0	92.61.191.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
92.114.16.0	92.114.19.255	Mobin Net Communication Company (Private Joint Stock)
92.114.20.0	92.114.21.255	Mobin Net Communication Company (Private Joint Stock)
92.114.22.0	92.114.23.255	Mobin Net Communication Company (Private Joint Stock)
92.114.24.0	92.114.24.255	Vira Cloud DataCenter PJSC
92.114.25.0	92.114.25.255	Mobin Net Communication Company (Private Joint Stock)
92.114.26.0	92.114.27.255	Mobin Net Communication Company (Private Joint Stock)
92.114.28.0	92.114.31.255	Mobin Net Communication Company (Private Joint Stock)
92.114.48.0	92.114.49.255	Cooperative Afra ertebatat-e-sabet-e Rasa Co
92.114.51.0	92.114.51.255	Tose'e Shabakeh Azad Ltd
92.114.64.0	92.114.79.255	Rightel Communication Service Company PJS
92.119.57.0	92.119.57.255	Negah Roshan Pars Company (PJS)
92.119.58.0	92.119.58.255	Negah Roshan Pars Company (PJS)
92.119.68.0	92.119.71.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
92.242.192.0	92.242.192.255	Respina Networks & Beyond PJSC
92.242.193.0	92.242.193.255	Respina Networks & Beyond PJSC
92.242.194.0	92.242.194.255	Respina Networks & Beyond PJSC
92.242.195.0	92.242.195.255	Respina Networks & Beyond PJSC
92.242.196.0	92.242.196.255	Respina Networks & Beyond PJSC
92.242.197.0	92.242.197.255	Respina Networks & Beyond PJSC
92.242.198.0	92.242.198.255	Respina Networks & Beyond PJSC
92.242.199.0	92.242.199.255	Respina Networks & Beyond PJSC
92.242.200.0	92.242.200.255	Respina Networks & Beyond PJSC
92.242.201.0	92.242.201.255	Respina Networks & Beyond PJSC
92.242.202.0	92.242.202.255	Respina Networks & Beyond PJSC
92.242.203.0	92.242.203.255	Respina Networks & Beyond PJSC
92.242.204.0	92.242.204.255	Respina Networks & Beyond PJSC
92.242.205.0	92.242.205.255	Respina Networks & Beyond PJSC
92.242.206.0	92.242.206.255	Respina Networks & Beyond PJSC
92.242.207.0	92.242.207.255	Respina Networks & Beyond PJSC
92.242.208.0	92.242.209.255	Respina Networks & Beyond PJSC
92.242.210.0	92.242.210.255	Respina Networks & Beyond PJSC
92.242.211.0	92.242.211.255	Respina Networks & Beyond PJSC
92.242.212.0	92.242.213.255	Respina Networks & Beyond PJSC
92.242.214.0	92.242.214.255	Respina Networks & Beyond PJSC
92.242.215.0	92.242.215.255	Respina Networks & Beyond PJSC
92.242.216.0	92.242.216.255	Respina Networks & Beyond PJSC
92.242.217.0	92.242.217.255	Respina Networks & Beyond PJSC
92.242.218.0	92.242.219.255	Respina Networks & Beyond PJSC
92.242.220.0	92.242.220.255	Respina Networks & Beyond PJSC
92.242.221.0	92.242.221.255	Respina Networks & Beyond PJSC
92.242.222.0	92.242.222.255	Respina Networks & Beyond PJSC
92.242.223.0	92.242.223.255	Respina Networks & Beyond PJSC
92.246.144.0	92.246.147.255	Iran Telecommunication Company Pjs
92.246.156.0	92.246.159.255	Iran Telecommunication Company Pjs
92.249.56.0	92.249.56.255	Sangan Khorasan Steel Mineral Industry Company (JSC)
92.249.57.0	92.249.57.255	Isfahan University of Technology
92.249.58.0	92.249.58.255	Isfahan University of Technology
92.249.59.0	92.249.59.255	Isfahan University of Technology
93.88.64.0	93.88.67.255	Iran Telecommunication Company Pjs
93.88.68.0	93.88.71.255	Iran Telecommunication Company Pjs
93.88.72.0	93.88.73.255	Iran Telecommunication Company Pjs
93.93.204.0	93.93.204.255	Middle East Mines and Mineral Industries Development Holding PLC
93.95.27.0	93.95.27.255	Netafraz Iranian Ltd.
93.110.0.0	93.110.15.255	Mobile Communication Company of Iran Plc
93.110.16.0	93.110.31.255	Mobile Communication Company of Iran Plc
93.110.32.0	93.110.47.255	Mobile Communication Company of Iran Plc
93.110.48.0	93.110.63.255	Mobile Communication Company of Iran Plc
93.110.64.0	93.110.79.255	Mobile Communication Company of Iran Plc
93.110.80.0	93.110.95.255	Mobile Communication Company of Iran Plc
93.110.96.0	93.110.111.255	Mobile Communication Company of Iran Plc
93.110.112.0	93.110.127.255	Mobile Communication Company of Iran Plc
93.110.128.0	93.110.143.255	Mobile Communication Company of Iran Plc
93.110.144.0	93.110.159.255	Mobile Communication Company of Iran Plc
93.110.160.0	93.110.175.255	Mobile Communication Company of Iran Plc
93.110.176.0	93.110.191.255	Mobile Communication Company of Iran Plc
93.110.192.0	93.110.207.255	Mobile Communication Company of Iran Plc
93.110.208.0	93.110.223.255	Mobile Communication Company of Iran Plc
93.110.224.0	93.110.239.255	Mobile Communication Company of Iran Plc
93.110.240.0	93.110.255.255	Mobile Communication Company of Iran Plc
93.113.224.0	93.113.224.255	Afranet
93.113.225.0	93.113.225.255	Afranet
93.113.226.0	93.113.226.255	Afranet
93.113.227.0	93.113.227.255	Afranet
93.113.228.0	93.113.228.255	Afranet
93.113.229.0	93.113.229.255	Afranet
93.113.230.0	93.113.230.255	Afranet
93.113.231.0	93.113.231.255	Afranet
93.113.232.0	93.113.232.255	Afranet
93.113.233.0	93.113.233.255	Afranet
93.113.234.0	93.113.234.255	Afranet
93.113.235.0	93.113.235.255	Afranet
93.113.236.0	93.113.236.255	Afranet
93.113.237.0	93.113.237.255	Afranet
93.113.238.0	93.113.238.255	Afranet
93.113.239.0	93.113.239.255	Afranet
93.114.16.0	93.114.31.255	Rightel Communication Service Company PJS
93.114.104.0	93.114.111.255	Iran Telecommunication Company Pjs
93.115.120.0	93.115.127.255	Iran Telecommunication Company Pjs
93.115.144.0	93.115.147.255	Asiatech Data Transmission Company
93.115.148.0	93.115.149.255	Asiatech Data Transmission Company
93.115.150.0	93.115.150.255	Asiatech Data Transmission Company
93.115.151.0	93.115.151.255	Asiatech Data Transmission Company
93.115.216.0	93.115.223.255	Iran Telecommunication Company Pjs
93.115.224.0	93.115.239.255	Iran Telecommunication Company Pjs
93.117.0.0	93.117.31.255	Iran Telecommunication Company Pjs
93.117.32.0	93.117.47.255	Iran Telecommunication Company Pjs
93.117.96.0	93.117.111.255	Iran Telecommunication Company Pjs
93.117.112.0	93.117.119.255	Iran Telecommunication Company Pjs
93.117.120.0	93.117.127.255	Iran Telecommunication Company Pjs
93.117.176.0	93.117.176.255	Mobile Communication Company of Iran Plc
93.117.177.0	93.117.177.255	Mobile Communication Company of Iran Plc
93.117.178.0	93.117.179.255	Mobile Communication Company of Iran Plc
93.117.180.0	93.117.183.255	Mobile Communication Company of Iran Plc
93.117.184.0	93.117.191.255	Mobile Communication Company of Iran Plc
93.118.96.0	93.118.103.255	Iran Telecommunication Company Pjs
93.118.104.0	93.118.107.255	Iran Telecommunication Company Pjs
93.118.108.0	93.118.109.255	Iran Telecommunication Company Pjs
93.118.110.0	93.118.111.255	Iran Telecommunication Company Pjs
93.118.112.0	93.118.127.255	Iran Telecommunication Company Pjs
93.118.128.0	93.118.159.255	Iran Telecommunication Company Pjs
93.118.160.0	93.118.175.255	Iran Telecommunication Company Pjs
93.118.180.0	93.118.181.255	Asiatech Data Transmission Company
93.118.182.0	93.118.183.255	Asiatech Data Transmission Company
93.118.184.0	93.118.187.255	Iran Telecommunication Company Pjs
93.119.32.0	93.119.63.255	Iran Telecommunication Company Pjs
93.119.64.0	93.119.95.255	Iran Telecommunication Company Pjs
93.119.208.0	93.119.211.255	Mobin Net Communication Company (Private Joint Stock)
93.119.212.0	93.119.215.255	Mobin Net Communication Company (Private Joint Stock)
93.119.216.0	93.119.219.255	Mobin Net Communication Company (Private Joint Stock)
93.119.220.0	93.119.223.255	Mobin Net Communication Company (Private Joint Stock)
93.126.0.0	93.126.7.255	Asmanfaraz Sepahan Company (P.J.S)
93.126.8.0	93.126.8.255	Asmanfaraz Sepahan Company (P.J.S)
93.126.9.0	93.126.9.255	Ertebatat Sabet Parsian Co. PJS
93.126.10.0	93.126.10.255	Foojan Cloud Computing Co. PJS
93.126.11.0	93.126.11.255	Asmanfaraz Sepahan Company (P.J.S)
93.126.12.0	93.126.12.255	Asmanfaraz Sepahan Company (P.J.S)
93.126.13.0	93.126.13.255	Tose'eh Lian Yas Ertebat PJSC
93.126.14.0	93.126.14.255	Azarakhsh Ava-E Ahvaz Co
93.126.15.0	93.126.15.255	Asmanfaraz Sepahan Company (P.J.S)
93.126.16.0	93.126.16.255	Lesun Communication Furtherance Engineers Co (Ltd.)
93.126.17.0	93.126.17.255	Asmanfaraz Sepahan Company (P.J.S)
93.126.18.0	93.126.19.255	Asmanfaraz Sepahan Company (P.J.S)
93.126.20.0	93.126.21.255	Asmanfaraz Sepahan Company (P.J.S)
93.126.22.0	93.126.22.255	Asmanfaraz Sepahan Company (P.J.S)
93.126.23.0	93.126.23.255	Azarakhsh Ava-E Ahvaz Co
93.126.24.0	93.126.25.255	Asmanfaraz Sepahan Company (P.J.S)
93.126.26.0	93.126.26.255	Asmanfaraz Sepahan Company (P.J.S)
93.126.27.0	93.126.27.255	Lesun Communication Furtherance Engineers Co (Ltd.)
93.126.28.0	93.126.29.255	Asmanfaraz Sepahan Company (P.J.S)
93.126.30.0	93.126.30.255	Azarakhsh Ava-E Ahvaz Co
93.126.31.0	93.126.31.255	Azarakhsh Ava-E Ahvaz Co
93.126.32.0	93.126.39.255	Asmanfaraz Sepahan Company (P.J.S)
93.126.40.0	93.126.43.255	Asmanfaraz Sepahan Company (P.J.S)
93.126.44.0	93.126.44.255	Asmanfaraz Sepahan Company (P.J.S)
93.126.45.0	93.126.45.255	Azarakhsh Ava-E Ahvaz Co
93.126.46.0	93.126.47.255	Asmanfaraz Sepahan Company (P.J.S)
93.126.48.0	93.126.51.255	Asmanfaraz Sepahan Company (P.J.S)
93.126.52.0	93.126.52.255	Lesun Communication Furtherance Engineers Co (Ltd.)
93.126.53.0	93.126.53.255	Asmanfaraz Sepahan Company (P.J.S)
93.126.54.0	93.126.55.255	Asmanfaraz Sepahan Company (P.J.S)
93.126.56.0	93.126.56.255	Asmanfaraz Sepahan Company (P.J.S)
93.126.57.0	93.126.57.255	Lesun Communication Furtherance Engineers Co (Ltd.)
93.126.58.0	93.126.58.255	Lesun Communication Furtherance Engineers Co (Ltd.)
93.126.59.0	93.126.59.255	Asmanfaraz Sepahan Company (P.J.S)
93.126.60.0	93.126.63.255	Asmanfaraz Sepahan Company (P.J.S)
93.190.24.0	93.190.25.255	Press TV Co.
93.190.26.0	93.190.27.255	Press TV Co.
93.190.28.0	93.190.31.255	Press TV Co.
94.24.0.0	94.24.15.255	Rightel Communication Service Company PJS
94.24.16.0	94.24.23.255	Rightel Communication Service Company PJS
94.24.80.0	94.24.95.255	Rightel Communication Service Company PJS
94.24.96.0	94.24.103.255	Rightel Communication Service Company PJS
94.74.128.0	94.74.129.255	Farahoosh Dena PLC
94.74.130.0	94.74.131.255	Farahoosh Dena PLC
94.74.132.0	94.74.135.255	Farahoosh Dena PLC
94.74.136.0	94.74.136.255	Farahoosh Dena PLC
94.74.138.0	94.74.139.255	Farahoosh Dena PLC
94.74.141.0	94.74.141.255	Farahoosh Dena PLC
94.74.142.0	94.74.143.255	Farahoosh Dena PLC
94.74.144.0	94.74.144.255	Farahoosh Dena PLC
94.74.145.0	94.74.145.255	Shabakeh Ertebatat Artak Towseeh PJSC
94.74.146.0	94.74.146.255	Farahoosh Dena PLC
94.74.148.0	94.74.149.255	Farahoosh Dena PLC
94.74.150.0	94.74.151.255	Farahoosh Dena PLC
94.74.152.0	94.74.155.255	Farahoosh Dena PLC
94.74.156.0	94.74.156.255	Farahoosh Dena PLC
94.74.160.0	94.74.163.255	Farahoosh Dena PLC
94.74.165.0	94.74.165.255	Farahoosh Dena PLC
94.74.166.0	94.74.167.255	Farahoosh Dena PLC
94.74.168.0	94.74.169.255	Farahoosh Dena PLC
94.74.170.0	94.74.170.255	Farahoosh Dena PLC
94.74.172.0	94.74.172.255	Farahoosh Dena PLC
94.74.173.0	94.74.173.255	Farahoosh Dena PLC
94.74.174.0	94.74.175.255	Farahoosh Dena PLC
94.74.176.0	94.74.179.255	Farahoosh Dena PLC
94.74.180.0	94.74.181.255	Farahoosh Dena PLC
94.74.182.0	94.74.183.255	Farahoosh Dena PLC
94.74.186.0	94.74.187.255	Farahoosh Dena PLC
94.74.188.0	94.74.189.255	Farahoosh Dena PLC
94.74.190.0	94.74.191.255	Farahoosh Dena PLC
94.101.128.0	94.101.135.255	Mobin Net Communication Company (Private Joint Stock)
94.101.136.0	94.101.136.255	Mobin Net Communication Company (Private Joint Stock)
94.101.137.0	94.101.137.255	Mobin Net Communication Company (Private Joint Stock)
94.101.138.0	94.101.138.255	Mobin Net Communication Company (Private Joint Stock)
94.101.139.0	94.101.139.255	Mobin Net Communication Company (Private Joint Stock)
94.101.140.0	94.101.141.255	Mobin Net Communication Company (Private Joint Stock)
94.101.142.0	94.101.142.255	Mobin Net Communication Company (Private Joint Stock)
94.101.143.0	94.101.143.255	Mobin Net Communication Company (Private Joint Stock)
94.101.176.0	94.101.179.255	Noyan Abr Arvan Co. ( Private Joint Stock)
94.101.180.0	94.101.181.255	Noyan Abr Arvan Co. ( Private Joint Stock)
94.101.182.0	94.101.182.255	Noyan Abr Arvan Co. ( Private Joint Stock)
94.101.183.0	94.101.183.255	Noyan Abr Arvan Co. ( Private Joint Stock)
94.101.184.0	94.101.187.255	Noyan Abr Arvan Co. ( Private Joint Stock)
94.101.188.0	94.101.188.255	Noyan Abr Arvan Co. ( Private Joint Stock)
94.101.189.0	94.101.189.255	Noyan Abr Arvan Co. ( Private Joint Stock)
94.101.190.0	94.101.190.255	Noyan Abr Arvan Co. ( Private Joint Stock)
94.101.191.0	94.101.191.255	Noyan Abr Arvan Co. ( Private Joint Stock)
94.101.240.0	94.101.240.255	Mobile Communication Company of Iran Plc
94.101.241.0	94.101.241.255	Mobile Communication Company of Iran Plc
94.101.242.0	94.101.243.255	Mobile Communication Company of Iran Plc
94.101.244.0	94.101.247.255	Mobile Communication Company of Iran Plc
94.101.248.0	94.101.255.255	Mobile Communication Company of Iran Plc
94.139.160.0	94.139.175.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
94.139.176.0	94.139.179.255	Datak Company LLC
94.139.180.0	94.139.180.255	Parvaresh Dadeha Co. Private Joint Stock
94.139.181.0	94.139.181.255	Parvaresh Dadeha Co. Private Joint Stock
94.139.182.0	94.139.182.255	Parvaresh Dadeha Co. Private Joint Stock
94.139.183.0	94.139.183.255	Parvaresh Dadeha Co. Private Joint Stock
94.139.184.0	94.139.184.255	Lesun Communication Furtherance Engineers Co (Ltd.)
94.139.185.0	94.139.185.255	Lesun Communication Furtherance Engineers Co (Ltd.)
94.139.186.0	94.139.186.255	Lesun Communication Furtherance Engineers Co (Ltd.)
94.139.187.0	94.139.187.255	Lesun Communication Furtherance Engineers Co (Ltd.)
94.139.188.0	94.139.188.255	Lesun Communication Furtherance Engineers Co (Ltd.)
94.139.189.0	94.139.189.255	Bank Refah Kargaran
94.139.190.0	94.139.191.255	Lesun Communication Furtherance Engineers Co (Ltd.)
94.176.8.0	94.176.15.255	Iran Telecommunication Company Pjs
94.176.32.0	94.176.39.255	Parvaresh Dadeha Co. Private Joint Stock
94.177.72.0	94.177.79.255	Parvaresh Dadeha Co. Private Joint Stock
94.182.0.0	94.182.7.255	Aria Shatel PJSC
94.182.8.0	94.182.9.255	Aria Shatel PJSC
94.182.10.0	94.182.10.255	Aria Shatel PJSC
94.182.11.0	94.182.11.255	Aria Shatel PJSC
94.182.12.0	94.182.13.255	Aria Shatel PJSC
94.182.14.0	94.182.14.255	Web Gostaran Bandar Company (PJS)
94.182.15.0	94.182.15.255	Aria Shatel PJSC
94.182.16.0	94.182.16.255	Aria Shatel PJSC
94.182.17.0	94.182.17.255	Pardis Fanvari Partak Ltd
94.182.18.0	94.182.19.255	Aria Shatel PJSC
94.182.20.0	94.182.23.255	Aria Shatel PJSC
94.182.24.0	94.182.27.255	Aria Shatel PJSC
94.182.28.0	94.182.29.255	Aria Shatel PJSC
94.182.30.0	94.182.31.255	Aria Shatel PJSC
94.182.32.0	94.182.35.255	Aria Shatel PJSC
94.182.36.0	94.182.36.255	Aria Shatel PJSC
94.182.37.0	94.182.37.255	Aria Shatel PJSC
94.182.38.0	94.182.38.255	Aria Shatel PJSC
94.182.39.0	94.182.39.255	Aria Shatel PJSC
94.182.40.0	94.182.41.255	Aria Shatel PJSC
94.182.42.0	94.182.42.255	Aria Shatel PJSC
94.182.43.0	94.182.43.255	Aria Shatel PJSC
94.182.44.0	94.182.44.255	Aria Shatel PJSC
94.182.45.0	94.182.45.255	Aria Shatel PJSC
94.182.46.0	94.182.46.255	Aria Shatel PJSC
94.182.47.0	94.182.47.255	Aria Shatel PJSC
94.182.48.0	94.182.55.255	Aria Shatel PJSC
94.182.56.0	94.182.56.255	Aria Shatel PJSC
94.182.57.0	94.182.57.255	Pardis Fanvari Partak Ltd
94.182.58.0	94.182.59.255	Aria Shatel PJSC
94.182.60.0	94.182.63.255	Aria Shatel PJSC
94.182.64.0	94.182.64.255	Aria Shatel PJSC
94.182.65.0	94.182.65.255	Aria Shatel PJSC
94.182.66.0	94.182.66.255	Aria Shatel PJSC
94.182.67.0	94.182.67.255	Aria Shatel PJSC
94.182.68.0	94.182.68.255	Aria Shatel PJSC
94.182.69.0	94.182.69.255	Aria Shatel PJSC
94.182.70.0	94.182.70.255	Aria Shatel PJSC
94.182.71.0	94.182.71.255	Aria Shatel PJSC
94.182.72.0	94.182.79.255	Aria Shatel PJSC
94.182.80.0	94.182.80.255	Aria Shatel PJSC
94.182.81.0	94.182.81.255	Aria Shatel PJSC
94.182.82.0	94.182.82.255	Aria Shatel PJSC
94.182.83.0	94.182.83.255	Aria Shatel PJSC
94.182.84.0	94.182.87.255	Aria Shatel PJSC
94.182.88.0	94.182.91.255	Aria Shatel PJSC
94.182.92.0	94.182.95.255	Aria Shatel PJSC
94.182.96.0	94.182.96.255	Aria Shatel PJSC
94.182.97.0	94.182.97.255	Aria Shatel PJSC
94.182.98.0	94.182.98.255	Aria Shatel PJSC
94.182.99.0	94.182.99.255	Aria Shatel PJSC
94.182.100.0	94.182.100.255	Aria Shatel PJSC
94.182.101.0	94.182.101.255	Aria Shatel PJSC
94.182.102.0	94.182.102.255	Aria Shatel PJSC
94.182.103.0	94.182.103.255	Aria Shatel PJSC
94.182.104.0	94.182.104.255	Fanavaran Mehr Vatan Tehran Server Group LLC
94.182.105.0	94.182.105.255	Fanavaran Mehr Vatan Tehran Server Group LLC
94.182.106.0	94.182.106.255	Aria Shatel PJSC
94.182.107.0	94.182.107.255	Fanavaran Mehr Vatan Tehran Server Group LLC
94.182.108.0	94.182.108.255	Fanavaran Mehr Vatan Tehran Server Group LLC
94.182.109.0	94.182.109.255	Fanavaran Mehr Vatan Tehran Server Group LLC
94.182.110.0	94.182.111.255	Aria Shatel PJSC
94.182.112.0	94.182.113.255	Aria Shatel PJSC
94.182.114.0	94.182.114.255	Fanavaran Mehr Vatan Tehran Server Group LLC
94.182.115.0	94.182.115.255	Aria Shatel PJSC
94.182.116.0	94.182.119.255	Aria Shatel PJSC
94.182.120.0	94.182.127.255	Aria Shatel PJSC
94.182.128.0	94.182.128.255	Aria Shatel PJSC
94.182.129.0	94.182.129.255	Fanavaran Mehr Vatan Tehran Server Group LLC
94.182.130.0	94.182.130.255	Aria Shatel PJSC
94.182.131.0	94.182.131.255	Fanavaran Mehr Vatan Tehran Server Group LLC
94.182.132.0	94.182.132.255	Fanavaran Mehr Vatan Tehran Server Group LLC
94.182.133.0	94.182.133.255	Aria Shatel PJSC
94.182.134.0	94.182.134.255	Fanavaran Mehr Vatan Tehran Server Group LLC
94.182.135.0	94.182.135.255	Fanavaran Mehr Vatan Tehran Server Group LLC
94.182.136.0	94.182.136.255	Fanavaran Mehr Vatan Tehran Server Group LLC
94.182.137.0	94.182.137.255	Aria Shatel PJSC
94.182.138.0	94.182.138.255	Aria Shatel PJSC
94.182.139.0	94.182.139.255	Aria Shatel PJSC
94.182.140.0	94.182.140.255	Aria Shatel PJSC
94.182.141.0	94.182.141.255	Aria Shatel PJSC
94.182.142.0	94.182.142.255	Aria Shatel PJSC
94.182.143.0	94.182.143.255	Aria Shatel PJSC
94.182.144.0	94.182.144.255	Fanavaran Mehr Vatan Tehran Server Group LLC
94.182.145.0	94.182.145.255	Fanavaran Mehr Vatan Tehran Server Group LLC
94.182.146.0	94.182.146.255	Aria Shatel PJSC
94.182.147.0	94.182.147.255	Fanavaran Mehr Vatan Tehran Server Group LLC
94.182.148.0	94.182.148.255	Aria Shatel PJSC
94.182.149.0	94.182.149.255	Fanavaran Mehr Vatan Tehran Server Group LLC
94.182.150.0	94.182.150.255	Fanavaran Mehr Vatan Tehran Server Group LLC
94.182.151.0	94.182.151.255	Fanavaran Mehr Vatan Tehran Server Group LLC
94.182.152.0	94.182.152.255	Fanavaran Mehr Vatan Tehran Server Group LLC
94.182.153.0	94.182.153.255	Aria Shatel PJSC
94.182.154.0	94.182.154.255	Aria Shatel PJSC
94.182.155.0	94.182.155.255	Aria Shatel PJSC
94.182.156.0	94.182.156.255	Aria Shatel PJSC
94.182.157.0	94.182.157.255	Aria Shatel PJSC
94.182.158.0	94.182.158.255	Aria Shatel PJSC
94.182.159.0	94.182.159.255	Aria Shatel PJSC
94.182.160.0	94.182.161.255	Aria Shatel PJSC
94.182.162.0	94.182.163.255	Aria Shatel PJSC
94.182.164.0	94.182.167.255	Aria Shatel PJSC
94.182.168.0	94.182.171.255	Aria Shatel PJSC
94.182.172.0	94.182.172.255	Aria Shatel PJSC
94.182.173.0	94.182.173.255	Aria Shatel PJSC
94.182.174.0	94.182.174.255	Aria Shatel PJSC
94.182.175.0	94.182.175.255	Aria Shatel PJSC
94.182.176.0	94.182.176.255	Aria Shatel PJSC
94.182.177.0	94.182.177.255	Aria Shatel PJSC
94.182.178.0	94.182.179.255	Aria Shatel PJSC
94.182.180.0	94.182.181.255	Aria Shatel PJSC
94.182.182.0	94.182.182.255	Aria Shatel PJSC
94.182.183.0	94.182.183.255	Aria Shatel PJSC
94.182.184.0	94.182.184.255	Aria Shatel PJSC
94.182.185.0	94.182.185.255	Aria Shatel PJSC
94.182.186.0	94.182.187.255	Aria Shatel PJSC
94.182.188.0	94.182.189.255	Aria Shatel PJSC
94.182.190.0	94.182.190.255	Aria Shatel PJSC
94.182.191.0	94.182.191.255	Aria Shatel PJSC
94.182.192.0	94.182.192.255	Aria Shatel PJSC
94.182.193.0	94.182.193.255	Aria Shatel PJSC
94.182.194.0	94.182.194.255	Aria Shatel PJSC
94.182.195.0	94.182.195.255	Avini Cultural and Art Institute
94.182.196.0	94.182.196.255	Aria Shatel PJSC
94.182.197.0	94.182.197.255	Aria Shatel PJSC
94.182.198.0	94.182.198.255	Aria Shatel PJSC
94.182.199.0	94.182.199.255	Aria Shatel PJSC
94.182.200.0	94.182.207.255	Aria Shatel PJSC
94.182.208.0	94.182.208.255	Aria Shatel PJSC
94.182.209.0	94.182.209.255	Aria Shatel PJSC
94.182.210.0	94.182.211.255	Aria Shatel PJSC
94.182.212.0	94.182.215.255	Aria Shatel PJSC
94.182.216.0	94.182.219.255	Aria Shatel PJSC
94.182.220.0	94.182.221.255	Aria Shatel PJSC
94.182.222.0	94.182.223.255	Aria Shatel PJSC
94.182.224.0	94.182.231.255	Aria Shatel PJSC
94.182.232.0	94.182.235.255	Aria Shatel PJSC
94.182.236.0	94.182.237.255	Aria Shatel PJSC
94.182.238.0	94.182.238.255	Aria Shatel PJSC
94.182.239.0	94.182.239.255	Aria Shatel PJSC
94.182.240.0	94.182.247.255	Aria Shatel PJSC
94.182.248.0	94.182.249.255	Aria Shatel PJSC
94.182.250.0	94.182.251.255	Aria Shatel PJSC
94.182.252.0	94.182.253.255	Aria Shatel PJSC
94.182.254.0	94.182.254.255	Pardis Fanvari Partak Ltd
94.182.255.0	94.182.255.255	Aria Shatel PJSC
94.183.0.0	94.183.7.255	Aria Shatel PJSC
94.183.8.0	94.183.11.255	Aria Shatel PJSC
94.183.12.0	94.183.12.255	Aria Shatel PJSC
94.183.13.0	94.183.13.255	Aria Shatel PJSC
94.183.14.0	94.183.15.255	Aria Shatel PJSC
94.183.16.0	94.183.16.255	Aria Shatel PJSC
94.183.17.0	94.183.17.255	Aria Shatel PJSC
94.183.18.0	94.183.18.255	Aria Shatel PJSC
94.183.19.0	94.183.19.255	Aria Shatel PJSC
94.183.20.0	94.183.21.255	Aria Shatel PJSC
94.183.22.0	94.183.22.255	Aria Shatel PJSC
94.183.23.0	94.183.23.255	Aria Shatel PJSC
94.183.24.0	94.183.31.255	Aria Shatel PJSC
94.183.32.0	94.183.39.255	Aria Shatel PJSC
94.183.40.0	94.183.47.255	Aria Shatel PJSC
94.183.48.0	94.183.55.255	Aria Shatel PJSC
94.183.56.0	94.183.57.255	Aria Shatel PJSC
94.183.58.0	94.183.58.255	Aria Shatel PJSC
94.183.59.0	94.183.59.255	Aria Shatel PJSC
94.183.60.0	94.183.63.255	Aria Shatel PJSC
94.183.64.0	94.183.79.255	Aria Shatel PJSC
94.183.80.0	94.183.95.255	Aria Shatel PJSC
94.183.96.0	94.183.111.255	Aria Shatel PJSC
94.183.112.0	94.183.127.255	Aria Shatel PJSC
94.183.128.0	94.183.135.255	Aria Shatel PJSC
94.183.136.0	94.183.143.255	Aria Shatel PJSC
94.183.144.0	94.183.144.255	Aria Shatel PJSC
94.183.145.0	94.183.145.255	Aria Shatel PJSC
94.183.146.0	94.183.146.255	Aria Shatel PJSC
94.183.147.0	94.183.147.255	Aria Shatel PJSC
94.183.148.0	94.183.148.255	Aria Shatel PJSC
94.183.149.0	94.183.149.255	Tahlil Dadeh Novin Fadak LLC
94.183.152.0	94.183.152.255	Aria Shatel PJSC
94.183.153.0	94.183.153.255	Giti Secure Cloud LLC
94.183.154.0	94.183.154.255	Giti Secure Cloud LLC
94.183.156.0	94.183.156.255	Mizban Gostar Dade Alvand Ltd
94.183.157.0	94.183.157.255	Mizban Gostar Dade Alvand Ltd
94.183.160.0	94.183.160.255	Mizban Gostar Dade Alvand Ltd
94.183.162.0	94.183.162.255	Giti Secure Cloud LLC
94.183.163.0	94.183.163.255	Toosee Ertebatat Damavand
94.183.164.0	94.183.164.255	Tahlil Dadeh Novin Fadak LLC
94.183.165.0	94.183.165.255	Ali Monfared
94.183.167.0	94.183.167.255	Pars Abr Toseeh Ertebatat Ltd
94.183.169.0	94.183.169.255	Giti Secure Cloud LLC
94.183.171.0	94.183.171.255	Seyed M. Ehsan MirGhafori
94.183.173.0	94.183.173.255	Nahor Hadish Design and Architecture Co. Ltd
94.183.175.0	94.183.175.255	Tejarat Electronic Nojhan Pardazesh LLC
94.183.176.0	94.183.176.255	Mizban Gostar Dade Alvand Ltd
94.183.179.0	94.183.179.255	Patron Technology Persia Ltd
94.183.180.0	94.183.180.255	Giti Secure Cloud LLC
94.183.181.0	94.183.181.255	Tahlil Dadeh Novin Fadak LLC
94.183.218.0	94.183.218.255	Giti Secure Cloud LLC
94.183.225.0	94.183.225.255	Mizban Gostar Dade Alvand Ltd
94.183.231.0	94.183.231.255	Tejarat Electronic Nojhan Pardazesh LLC
94.183.237.0	94.183.237.255	Mizbani Hooshmand Mehr Afarin PJSC
94.183.240.0	94.183.240.255	Milad Ahadpour
94.184.0.0	94.184.3.255	Institute for Research in Fundamental Sciences
94.184.4.0	94.184.4.255	Saeid Shahrokhi
94.184.5.0	94.184.5.255	Institute for Research in Fundamental Sciences
94.184.6.0	94.184.6.255	Institute for Research in Fundamental Sciences
94.184.7.0	94.184.7.255	Amir Hosein Maaref
94.184.8.0	94.184.8.255	Amir Hosein Maaref
94.184.9.0	94.184.9.255	Asiatech Data Transmission Company
94.184.10.0	94.184.10.255	Tejarat Electronic Nojhan Pardazesh LLC
94.184.11.0	94.184.11.255	Institute for Research in Fundamental Sciences
94.184.12.0	94.184.13.255	Institute for Research in Fundamental Sciences
94.184.14.0	94.184.14.255	Amir Hosein Maaref
94.184.15.0	94.184.15.255	Amir Hosein Maaref
94.184.16.0	94.184.16.255	Asiatech Data Transmission Company
94.184.17.0	94.184.17.255	Tejarat Electronic Nojhan Pardazesh LLC
94.184.18.0	94.184.18.255	Tejarat Electronic Nojhan Pardazesh LLC
94.184.19.0	94.184.19.255	Institute for Research in Fundamental Sciences
94.184.20.0	94.184.20.255	Institute for Research in Fundamental Sciences
94.184.21.0	94.184.21.255	Foojan Cloud Computing Co. PJS
94.184.22.0	94.184.23.255	Institute for Research in Fundamental Sciences
94.184.24.0	94.184.24.255	Institute for Research in Fundamental Sciences
94.184.25.0	94.184.25.255	Amir Hosein Maaref
94.184.26.0	94.184.26.255	Institute for Research in Fundamental Sciences
94.184.27.0	94.184.27.255	Amir Hosein Maaref
94.184.28.0	94.184.28.255	Institute for Research in Fundamental Sciences
94.184.29.0	94.184.29.255	Saeid Shahrokhi
94.184.30.0	94.184.31.255	Institute for Research in Fundamental Sciences
94.184.32.0	94.184.63.255	Institute for Research in Fundamental Sciences
94.184.64.0	94.184.79.255	Institute for Research in Fundamental Sciences
94.184.80.0	94.184.83.255	Institute for Research in Fundamental Sciences
94.184.84.0	94.184.87.255	Kharazmi University
94.184.88.0	94.184.91.255	University of Isfahan
94.184.92.0	94.184.92.255	Tebyan-e-Noor Cultural-Artistic Institute
94.184.93.0	94.184.93.255	Tebyan-e-Noor Cultural-Artistic Institute
94.184.94.0	94.184.95.255	Institute for Research in Fundamental Sciences
94.184.96.0	94.184.111.255	Tehran University of Medical Science
94.184.112.0	94.184.119.255	Shahid Beheshti University
94.184.120.0	94.184.127.255	Institute for Research in Fundamental Sciences
94.184.128.0	94.184.135.255	Institute for Research in Fundamental Sciences
94.184.136.0	94.184.139.255	Institute for Research in Fundamental Sciences
94.184.140.0	94.184.140.255	Informatics Services PJSC
94.184.141.0	94.184.141.255	Informatics Services PJSC
94.184.142.0	94.184.143.255	Informatics Services PJSC
94.184.144.0	94.184.159.255	Tehran University of Medical Science
94.184.160.0	94.184.175.255	Institute for Research in Fundamental Sciences
94.184.176.0	94.184.176.255	Kajeh Nasir Toosi University of Technology
94.184.177.0	94.184.177.255	Kajeh Nasir Toosi University of Technology
94.184.178.0	94.184.179.255	Kajeh Nasir Toosi University of Technology
94.184.180.0	94.184.183.255	Kajeh Nasir Toosi University of Technology
94.184.184.0	94.184.191.255	Institute for Research in Fundamental Sciences
94.184.192.0	94.184.255.255	Institute for Research in Fundamental Sciences
94.199.0.0	94.199.0.255	Cooperative Afra ertebatat-e-sabet-e Rasa Co
94.199.136.0	94.199.137.255	Isfahan University of Medical Sciences & Health Services
94.199.138.0	94.199.139.255	Isfahan University of Medical Sciences & Health Services
94.232.168.0	94.232.168.255	Tebyan-e-Noor Cultural-Artistic Institute
94.232.169.0	94.232.169.255	Tebyan-e-Noor Cultural-Artistic Institute
94.232.170.0	94.232.170.255	Tebyan-e-Noor Cultural-Artistic Institute
94.232.171.0	94.232.171.255	Tebyan-e-Noor Cultural-Artistic Institute
94.232.172.0	94.232.172.255	Tebyan-e-Noor Cultural-Artistic Institute
94.232.173.0	94.232.173.255	Tebyan-e-Noor Cultural-Artistic Institute
94.232.174.0	94.232.174.255	Tebyan-e-Noor Cultural-Artistic Institute
94.232.175.0	94.232.175.255	Tebyan-e-Noor Cultural-Artistic Institute
95.38.0.0	95.38.7.255	Fanava Group
95.38.8.0	95.38.9.255	Fanava Group
95.38.10.0	95.38.10.255	Fanava Group
95.38.11.0	95.38.11.255	Fanava Group
95.38.12.0	95.38.12.255	Fanava Group
95.38.13.0	95.38.13.255	Fanava Group
95.38.14.0	95.38.14.255	Fanava Group
95.38.15.0	95.38.15.255	Fanava Group
95.38.16.0	95.38.16.255	Fanava Group
95.38.17.0	95.38.17.255	Fanava Group
95.38.18.0	95.38.18.255	Fanava Group
95.38.19.0	95.38.19.255	Fanava Group
95.38.20.0	95.38.20.255	Fanava Group
95.38.21.0	95.38.21.255	Fanava Group
95.38.22.0	95.38.22.255	Fanava Group
95.38.23.0	95.38.23.255	Fanava Group
95.38.24.0	95.38.24.255	Fanava Group
95.38.25.0	95.38.25.255	Fanava Group
95.38.26.0	95.38.26.255	Fanava Group
95.38.27.0	95.38.27.255	Fanava Group
95.38.28.0	95.38.28.255	Fanava Group
95.38.29.0	95.38.29.255	Fanava Group
95.38.30.0	95.38.30.255	Fanava Group
95.38.31.0	95.38.31.255	Fanava Group
95.38.32.0	95.38.32.255	Fanava Group
95.38.33.0	95.38.33.255	Fanava Group
95.38.34.0	95.38.34.255	Fanava Group
95.38.35.0	95.38.35.255	Fanava Group
95.38.36.0	95.38.37.255	Fanava Group
95.38.38.0	95.38.38.255	Atrin Information & Communications Technology Company PJS
95.38.39.0	95.38.39.255	Fanava Group
95.38.40.0	95.38.40.255	Fanava Group
95.38.41.0	95.38.41.255	Fanava Group
95.38.42.0	95.38.42.255	Fanava Group
95.38.43.0	95.38.43.255	Fanava Group
95.38.44.0	95.38.44.255	Fanava Group
95.38.45.0	95.38.45.255	Dade Samane Fanava Company (PJS)
95.38.46.0	95.38.46.255	Fanava Group
95.38.47.0	95.38.47.255	Fanava Group
95.38.48.0	95.38.48.255	Fanava Group
95.38.49.0	95.38.49.255	Fanava Group
95.38.50.0	95.38.50.255	Fanava Group
95.38.51.0	95.38.51.255	Fanava Group
95.38.52.0	95.38.52.255	Fanava Group
95.38.53.0	95.38.53.255	Fanava Group
95.38.54.0	95.38.55.255	Dade Samane Fanava Company (PJS)
95.38.56.0	95.38.57.255	Fanava Group
95.38.58.0	95.38.58.255	Fanava Group
95.38.59.0	95.38.59.255	Fanava Group
95.38.60.0	95.38.60.255	Fanava Group
95.38.61.0	95.38.61.255	Fanava Group
95.38.62.0	95.38.62.255	Fanava Group
95.38.63.0	95.38.63.255	Fanava Group
95.38.64.0	95.38.64.255	Fanava Group
95.38.65.0	95.38.65.255	Fanava Group
95.38.66.0	95.38.67.255	Fanava Group
95.38.68.0	95.38.68.255	Fanava Group
95.38.69.0	95.38.69.255	Fanava Group
95.38.70.0	95.38.71.255	Fanava Group
95.38.72.0	95.38.72.255	Azarakhsh Ava-E Ahvaz Co
95.38.73.0	95.38.73.255	Fanava Group
95.38.74.0	95.38.74.255	Fanava Group
95.38.75.0	95.38.75.255	Fanava Group
95.38.76.0	95.38.76.255	Fanava Group
95.38.77.0	95.38.77.255	Fanava Group
95.38.78.0	95.38.78.255	Fanava Group
95.38.79.0	95.38.79.255	Fanava Group
95.38.80.0	95.38.81.255	Fanava Group
95.38.82.0	95.38.83.255	Fanava Group
95.38.84.0	95.38.85.255	Fanava Group
95.38.86.0	95.38.86.255	Fanava Group
95.38.87.0	95.38.87.255	Fanava Group
95.38.88.0	95.38.88.255	Fanava Group
95.38.89.0	95.38.89.255	Fanava Group
95.38.90.0	95.38.90.255	Fanava Group
95.38.91.0	95.38.91.255	Fanava Group
95.38.92.0	95.38.93.255	Fanava Group
95.38.94.0	95.38.95.255	Fanava Group
95.38.96.0	95.38.96.255	Respina Networks & Beyond PJSC
95.38.97.0	95.38.97.255	Respina Networks & Beyond PJSC
95.38.98.0	95.38.98.255	Respina Networks & Beyond PJSC
95.38.99.0	95.38.99.255	Respina Networks & Beyond PJSC
95.38.100.0	95.38.100.255	Respina Networks & Beyond PJSC
95.38.101.0	95.38.101.255	Respina Networks & Beyond PJSC
95.38.102.0	95.38.102.255	Respina Networks & Beyond PJSC
95.38.103.0	95.38.103.255	Respina Networks & Beyond PJSC
95.38.104.0	95.38.104.255	Respina Networks & Beyond PJSC
95.38.105.0	95.38.105.255	Respina Networks & Beyond PJSC
95.38.106.0	95.38.106.255	Respina Networks & Beyond PJSC
95.38.107.0	95.38.107.255	Respina Networks & Beyond PJSC
95.38.108.0	95.38.109.255	Fanava Group
95.38.110.0	95.38.111.255	Fanava Group
95.38.112.0	95.38.119.255	Fanava Group
95.38.120.0	95.38.121.255	Fanava Group
95.38.122.0	95.38.123.255	Fanava Group
95.38.124.0	95.38.124.255	Fanava Group
95.38.125.0	95.38.125.255	Fanava Group
95.38.126.0	95.38.127.255	Fanava Group
95.38.128.0	95.38.129.255	Fanava Group
95.38.130.0	95.38.130.255	Respina Networks & Beyond PJSC
95.38.131.0	95.38.131.255	Respina Networks & Beyond PJSC
95.38.132.0	95.38.132.255	Respina Networks & Beyond PJSC
95.38.133.0	95.38.133.255	Fanava Group
95.38.134.0	95.38.134.255	Respina Networks & Beyond PJSC
95.38.135.0	95.38.135.255	Respina Networks & Beyond PJSC
95.38.136.0	95.38.136.255	Fanava Group
95.38.137.0	95.38.137.255	Respina Networks & Beyond PJSC
95.38.138.0	95.38.138.255	Respina Networks & Beyond PJSC
95.38.139.0	95.38.139.255	Fanava Group
95.38.140.0	95.38.140.255	Fanava Group
95.38.141.0	95.38.141.255	Respina Networks & Beyond PJSC
95.38.142.0	95.38.142.255	Respina Networks & Beyond PJSC
95.38.143.0	95.38.143.255	Respina Networks & Beyond PJSC
95.38.144.0	95.38.147.255	Cooperative Afra ertebatat-e-sabet-e Rasa Co
95.38.148.0	95.38.149.255	Cooperative Afra ertebatat-e-sabet-e Rasa Co
95.38.150.0	95.38.150.255	Cooperative Afra ertebatat-e-sabet-e Rasa Co
95.38.151.0	95.38.151.255	Fanava Group
95.38.152.0	95.38.152.255	Respina Networks & Beyond PJSC
95.38.153.0	95.38.153.255	Respina Networks & Beyond PJSC
95.38.154.0	95.38.154.255	Respina Networks & Beyond PJSC
95.38.155.0	95.38.155.255	Respina Networks & Beyond PJSC
95.38.156.0	95.38.156.255	Respina Networks & Beyond PJSC
95.38.157.0	95.38.157.255	Respina Networks & Beyond PJSC
95.38.158.0	95.38.158.255	Respina Networks & Beyond PJSC
95.38.159.0	95.38.159.255	Respina Networks & Beyond PJSC
95.38.160.0	95.38.161.255	Noyan Abr Arvan Co. ( Private Joint Stock)
95.38.162.0	95.38.163.255	Noyan Abr Arvan Co. ( Private Joint Stock)
95.38.164.0	95.38.165.255	Noyan Abr Arvan Co. ( Private Joint Stock)
95.38.166.0	95.38.167.255	Noyan Abr Arvan Co. ( Private Joint Stock)
95.38.168.0	95.38.168.255	Respina Networks & Beyond PJSC
95.38.169.0	95.38.169.255	Respina Networks & Beyond PJSC
95.38.170.0	95.38.170.255	Respina Networks & Beyond PJSC
95.38.171.0	95.38.171.255	Respina Networks & Beyond PJSC
95.38.172.0	95.38.173.255	Fanava Group
95.38.174.0	95.38.174.255	Respina Networks & Beyond PJSC
95.38.175.0	95.38.175.255	Fanava Group
95.38.176.0	95.38.177.255	Noyan Abr Arvan Co. ( Private Joint Stock)
95.38.178.0	95.38.179.255	Noyan Abr Arvan Co. ( Private Joint Stock)
95.38.180.0	95.38.181.255	Noyan Abr Arvan Co. ( Private Joint Stock)
95.38.182.0	95.38.183.255	Noyan Abr Arvan Co. ( Private Joint Stock)
95.38.184.0	95.38.185.255	Noyan Abr Arvan Co. ( Private Joint Stock)
95.38.186.0	95.38.187.255	Noyan Abr Arvan Co. ( Private Joint Stock)
95.38.188.0	95.38.189.255	Noyan Abr Arvan Co. ( Private Joint Stock)
95.38.190.0	95.38.191.255	Noyan Abr Arvan Co. ( Private Joint Stock)
95.38.192.0	95.38.192.255	Fanava Group
95.38.193.0	95.38.193.255	Fanava Group
95.38.194.0	95.38.194.255	Fanava Group
95.38.195.0	95.38.195.255	Atrin Information & Communications Technology Company PJS
95.38.196.0	95.38.196.255	Fanava Group
95.38.197.0	95.38.197.255	Atrin Information & Communications Technology Company PJS
95.38.198.0	95.38.198.255	Peimayesh Ertebat Oxin Co.(PJSC)
95.38.199.0	95.38.199.255	Fanava Group
95.38.200.0	95.38.200.255	Respina Networks & Beyond PJSC
95.38.201.0	95.38.201.255	Respina Networks & Beyond PJSC
95.38.202.0	95.38.202.255	Respina Networks & Beyond PJSC
95.38.203.0	95.38.203.255	Respina Networks & Beyond PJSC
95.38.204.0	95.38.207.255	Fanava Group
95.38.208.0	95.38.208.255	Fanava Group
95.38.209.0	95.38.209.255	Respina Networks & Beyond PJSC
95.38.210.0	95.38.210.255	Fanava Group
95.38.211.0	95.38.211.255	Respina Networks & Beyond PJSC
95.38.212.0	95.38.215.255	Fanava Group
95.38.216.0	95.38.219.255	Fanava Group
95.38.220.0	95.38.223.255	Fanava Group
95.38.224.0	95.38.224.255	Fanava Group
95.38.225.0	95.38.225.255	Fanava Group
95.38.226.0	95.38.227.255	Fanava Group
95.38.228.0	95.38.231.255	Fanava Group
95.38.232.0	95.38.233.255	Noyan Abr Arvan Co. ( Private Joint Stock)
95.38.234.0	95.38.235.255	Fanava Group
95.38.236.0	95.38.239.255	Fanava Group
95.38.240.0	95.38.240.255	Respina Networks & Beyond PJSC
95.38.241.0	95.38.241.255	Respina Networks & Beyond PJSC
95.38.242.0	95.38.242.255	Respina Networks & Beyond PJSC
95.38.243.0	95.38.243.255	Respina Networks & Beyond PJSC
95.38.244.0	95.38.244.255	Respina Networks & Beyond PJSC
95.38.245.0	95.38.245.255	Respina Networks & Beyond PJSC
95.38.246.0	95.38.246.255	Respina Networks & Beyond PJSC
95.38.247.0	95.38.247.255	Respina Networks & Beyond PJSC
95.38.248.0	95.38.248.255	Respina Networks & Beyond PJSC
95.38.249.0	95.38.249.255	Respina Networks & Beyond PJSC
95.38.250.0	95.38.250.255	Respina Networks & Beyond PJSC
95.38.251.0	95.38.251.255	Respina Networks & Beyond PJSC
95.38.252.0	95.38.252.255	Fanava Group
95.38.253.0	95.38.253.255	Fanava Group
95.38.254.0	95.38.254.255	Fanava Group
95.38.255.0	95.38.255.255	Fanava Group
95.64.0.0	95.64.15.255	Mobile Communication Company of Iran Plc
95.64.16.0	95.64.31.255	Mobile Communication Company of Iran Plc
95.64.32.0	95.64.47.255	Mobile Communication Company of Iran Plc
95.64.48.0	95.64.63.255	Mobile Communication Company of Iran Plc
95.64.64.0	95.64.79.255	Mobile Communication Company of Iran Plc
95.64.80.0	95.64.95.255	Mobile Communication Company of Iran Plc
95.64.96.0	95.64.111.255	Mobile Communication Company of Iran Plc
95.64.112.0	95.64.127.255	Mobile Communication Company of Iran Plc
95.80.128.0	95.80.135.255	Respina Networks & Beyond PJSC
95.80.136.0	95.80.139.255	Bozorg Net-E Aria
95.80.140.0	95.80.143.255	Bozorg Net-E Aria
95.80.144.0	95.80.159.255	Bozorg Net-E Aria
95.80.160.0	95.80.167.255	Bozorg Net-E Aria
95.80.168.0	95.80.171.255	Bozorg Net-E Aria
95.80.172.0	95.80.175.255	Bozorg Net-E Aria
95.80.176.0	95.80.183.255	Bozorg Net-E Aria
95.80.184.0	95.80.185.255	Bozorg Net-E Aria
95.80.186.0	95.80.186.255	Bozorg Net-E Aria
95.80.187.0	95.80.187.255	Bozorg Net-E Aria
95.80.188.0	95.80.191.255	Bozorg Net-E Aria
95.81.78.0	95.81.78.255	Andishe Sabz Khazar Co. P.J.S.
95.128.159.0	95.128.159.255	Mabna Financial Data Processing PJS
95.128.194.0	95.128.194.255	Dade Pardazi Mobinhost Co Ltd
95.130.56.0	95.130.56.255	Khalij Fars Ettela Resan LTD
95.130.57.0	95.130.57.255	Khalij Fars Ettela Resan LTD
95.130.58.0	95.130.58.255	Khalij Fars Ettela Resan LTD
95.130.59.0	95.130.59.255	Khalij Fars Ettela Resan LTD
95.130.60.0	95.130.60.255	Khalij Fars Ettela Resan LTD
95.130.61.0	95.130.61.255	Khalij Fars Ettela Resan LTD
95.130.62.0	95.130.62.255	Khalij Fars Ettela Resan LTD
95.130.63.0	95.130.63.255	Khalij Fars Ettela Resan LTD
95.130.225.0	95.130.225.255	Pooya Parto Qeshm Cooperative Company
95.130.240.0	95.130.240.255	Atieh Dadeh Pardaz PJSC
95.130.241.0	95.130.241.255	Atieh Dadeh Pardaz PJSC
95.130.242.0	95.130.242.255	Atieh Dadeh Pardaz PJSC
95.130.243.0	95.130.243.255	Atieh Dadeh Pardaz PJSC
95.130.244.0	95.130.244.255	Atieh Dadeh Pardaz PJSC
95.130.245.0	95.130.245.255	Atieh Dadeh Pardaz PJSC
95.130.246.0	95.130.246.255	Atieh Dadeh Pardaz PJSC
95.130.247.0	95.130.247.255	Atieh Dadeh Pardaz PJSC
95.142.224.0	95.142.224.255	Armaghan Rahe Talaie
95.142.225.0	95.142.225.255	Armaghan Rahe Talaie
95.142.226.0	95.142.226.255	Armaghan Rahe Talaie
95.142.227.0	95.142.227.255	Armaghan Rahe Talaie
95.142.228.0	95.142.228.255	Armaghan Rahe Talaie
95.142.229.0	95.142.229.255	Armaghan Rahe Talaie
95.142.230.0	95.142.230.255	Armaghan Rahe Talaie
95.142.231.0	95.142.231.255	Farzanegan Pars Communications Company PJS
95.142.232.0	95.142.232.255	Farzanegan Pars Communications Company PJS
95.142.233.0	95.142.233.255	Armaghan Rahe Talaie
95.142.234.0	95.142.234.255	Armaghan Rahe Talaie
95.142.235.0	95.142.235.255	Armaghan Rahe Talaie
95.142.236.0	95.142.236.255	Armaghan Rahe Talaie
95.142.237.0	95.142.237.255	Armaghan Rahe Talaie
95.142.238.0	95.142.238.255	Farzanegan Pars Communications Company PJS
95.142.239.0	95.142.239.255	Farzanegan Pars Communications Company PJS
95.156.222.0	95.156.223.255	Parsan Lin Co. PJS
95.156.233.0	95.156.233.255	Parsan Lin Co. PJS
95.156.234.0	95.156.235.255	Parsan Lin Co. PJS
95.156.236.0	95.156.236.255	Parsan Lin Co. PJS
95.156.237.0	95.156.237.255	Parsan Lin Co. PJS
95.156.248.0	95.156.249.255	Parsan Lin Co. PJS
95.156.252.0	95.156.252.255	Parsan Lin Co. PJS
95.156.253.0	95.156.253.255	Parsan Lin Co. PJS
95.156.254.0	95.156.254.255	Parsan Lin Co. PJS
95.156.255.0	95.156.255.255	Parsan Lin Co. PJS
95.162.0.0	95.162.31.255	Rightel Communication Service Company PJS
95.162.32.0	95.162.63.255	Rightel Communication Service Company PJS
95.162.64.0	95.162.127.255	Rightel Communication Service Company PJS
95.162.128.0	95.162.159.255	Rightel Communication Service Company PJS
95.162.160.0	95.162.175.255	Rightel Communication Service Company PJS
95.162.176.0	95.162.183.255	Rightel Communication Service Company PJS
95.162.184.0	95.162.191.255	Rightel Communication Service Company PJS
95.162.192.0	95.162.223.255	Rightel Communication Service Company PJS
95.162.224.0	95.162.239.255	Rightel Communication Service Company PJS
95.162.240.0	95.162.255.255	Rightel Communication Service Company PJS
95.215.59.0	95.215.59.255	Paya Bastar Arina Co. Private J.S.
95.215.160.0	95.215.161.255	Gostaresh Ertebat Azin Kia Company PJSC
95.215.162.0	95.215.163.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
103.111.69.0	103.111.69.255	Mobin Net Communication Company (Private Joint Stock)
103.111.71.0	103.111.71.255	Mobin Net Communication Company (Private Joint Stock)
103.130.144.0	103.130.144.255	Ayande Sazan Pardazesh AMN Ltd.
103.130.146.0	103.130.146.255	Faraso Samaneh Pasargad Ltd.
103.130.147.0	103.130.147.255	Computer Research Center of Islamic Sciences
103.132.228.0	103.132.229.255	Mobin Net Communication Company (Private Joint Stock)
103.140.128.0	103.140.129.255	Mobin Net Communication Company (Private Joint Stock)
103.215.220.0	103.215.220.255	Web Gostaran Bandar Company (PJS)
103.215.221.0	103.215.221.255	Web Gostaran Bandar Company (PJS)
103.215.222.0	103.215.222.255	Sindad Network Technology PJSC
103.215.223.0	103.215.223.255	Web Gostaran Bandar Company (PJS)
103.216.60.0	103.216.60.255	Sindad Network Technology PJSC
103.216.61.0	103.216.61.255	Sindad Network Technology PJSC
103.216.62.0	103.216.62.255	Sindad Network Technology PJSC
103.216.63.0	103.216.63.255	Sindad Network Technology PJSC
103.217.124.0	103.217.127.255	Mobin Net Communication Company (Private Joint Stock)
103.231.136.0	103.231.136.255	Fara Negar Pardaz Noor Khuzestan Co.JSP
103.231.137.0	103.231.137.255	Fara Negar Pardaz Noor Khuzestan Co.JSP
103.231.138.0	103.231.138.255	Fara Negar Pardaz Noor Khuzestan Co.JSP
103.231.139.0	103.231.139.255	Fara Negar Pardaz Noor Khuzestan Co.JSP
109.70.76.0	109.70.76.255	Web Dadeh Paydar Co (Ltd)
109.70.77.0	109.70.77.255	Web Dadeh Paydar Co (Ltd)
109.70.78.0	109.70.78.255	Web Dadeh Paydar Co (Ltd)
109.70.237.0	109.70.237.255	Enteghal Dadeh Mahan Co. PJSC
109.72.192.0	109.72.199.255	Khalij Fars Ettela Resan LTD
109.72.200.0	109.72.200.255	Khalij Fars Ettela Resan LTD
109.72.201.0	109.72.201.255	Khalij Fars Ettela Resan LTD
109.72.202.0	109.72.203.255	Khalij Fars Ettela Resan LTD
109.72.204.0	109.72.207.255	Khalij Fars Ettela Resan LTD
109.74.224.0	109.74.239.255	Iran Telecommunication Company Pjs
109.94.164.0	109.94.164.255	Farhang Azma Communications Company LTD
109.94.165.0	109.94.165.255	Parmin Cloud Computing LLC
109.94.166.0	109.94.166.255	Tadbir Pardaz IT Group LTD
109.94.167.0	109.94.167.255	Tadbir Pardaz IT Group LTD
109.95.60.0	109.95.60.255	Imen Sanat Novin Alighapoo Private Joint Stock Company
109.95.61.0	109.95.61.255	Imen Sanat Novin Alighapoo Private Joint Stock Company
109.95.62.0	109.95.62.255	Gostaresh Ertebat Azin Kia Company PJSC
109.95.63.0	109.95.63.255	Pouya Nazm Najafabad Computer Co(PJS)
109.95.64.0	109.95.64.255	Guilanet
109.95.69.0	109.95.69.255	Guilanet
109.95.71.0	109.95.71.255	Guilanet
109.107.131.0	109.107.131.255	Omid Entrepreneurship Fund PLC
109.108.160.0	109.108.160.255	Mobile Communication Company of Iran Plc
109.108.161.0	109.108.161.255	Mobile Communication Company of Iran Plc
109.108.162.0	109.108.163.255	Mobile Communication Company of Iran Plc
109.108.164.0	109.108.167.255	Mobile Communication Company of Iran Plc
109.108.168.0	109.108.175.255	Mobile Communication Company of Iran Plc
109.108.176.0	109.108.191.255	Mobile Communication Company of Iran Plc
109.109.32.0	109.109.63.255	Andishe Sabz Khazar Co. P.J.S.
109.122.224.0	109.122.231.255	Asiatech Data Transmission Company
109.122.232.0	109.122.239.255	Asiatech Data Transmission Company
109.122.240.0	109.122.240.255	Avini Cultural and Art Institute
109.122.241.0	109.122.241.255	Respina Networks & Beyond PJSC
109.122.242.0	109.122.242.255	Respina Networks & Beyond PJSC
109.122.243.0	109.122.243.255	Respina Networks & Beyond PJSC
109.122.244.0	109.122.244.255	Amir Mohammad Masoudi
109.122.245.0	109.122.245.255	Hamed Shirvani Bagheri
109.122.246.0	109.122.246.255	Vandad Vira Hooman LLC
109.122.247.0	109.122.247.255	Vandad Vira Hooman LLC
109.122.248.0	109.122.248.255	Respina Networks & Beyond PJSC
109.122.249.0	109.122.249.255	Vandad Vira Hooman LLC
109.122.250.0	109.122.250.255	Vandad Vira Hooman LLC
109.122.251.0	109.122.251.255	Mizbani Hooshmand Mehr Afarin PJSC
109.122.252.0	109.122.252.255	Vandad Vira Hooman LLC
109.122.253.0	109.122.253.255	Respina Networks & Beyond PJSC
109.122.254.0	109.122.254.255	FanAvaran Mihan Mizban PJSC
109.125.128.0	109.125.131.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
109.125.132.0	109.125.135.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
109.125.136.0	109.125.143.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
109.125.144.0	109.125.159.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
109.125.160.0	109.125.167.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
109.125.168.0	109.125.175.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
109.125.176.0	109.125.179.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
109.125.180.0	109.125.183.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
109.125.184.0	109.125.184.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
109.125.185.0	109.125.185.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
109.125.186.0	109.125.187.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
109.125.188.0	109.125.189.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
109.125.190.0	109.125.190.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
109.125.191.0	109.125.191.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
109.162.128.0	109.162.143.255	Datak Company LLC
109.162.144.0	109.162.159.255	Datak Company LLC
109.162.160.0	109.162.175.255	Datak Company LLC
109.162.176.0	109.162.191.255	Datak Company LLC
109.162.192.0	109.162.207.255	Datak Company LLC
109.162.208.0	109.162.223.255	Datak Company LLC
109.162.224.0	109.162.239.255	Datak Company LLC
109.162.240.0	109.162.247.255	Kish Cell Pars Co. (PJS)
109.162.248.0	109.162.249.255	Datak Company LLC
109.162.250.0	109.162.250.255	Atrin Information & Communications Technology Company PJS
109.162.251.0	109.162.251.255	Atrin Information & Communications Technology Company PJS
109.162.252.0	109.162.252.255	Atrin Information & Communications Technology Company PJS
109.162.253.0	109.162.253.255	Datak Company LLC
109.162.254.0	109.162.255.255	Datak Company LLC
109.201.0.0	109.201.0.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
109.201.1.0	109.201.1.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
109.201.2.0	109.201.2.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
109.201.3.0	109.201.3.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
109.201.4.0	109.201.7.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
109.201.8.0	109.201.9.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
109.201.10.0	109.201.10.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
109.201.11.0	109.201.11.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
109.201.12.0	109.201.13.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
109.201.14.0	109.201.14.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
109.201.15.0	109.201.15.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
109.201.16.0	109.201.17.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
109.201.18.0	109.201.18.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
109.201.19.0	109.201.19.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
109.201.20.0	109.201.20.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
109.201.21.0	109.201.21.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
109.201.22.0	109.201.23.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
109.201.24.0	109.201.25.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
109.201.26.0	109.201.26.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
109.201.27.0	109.201.27.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
109.201.28.0	109.201.31.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
109.203.128.0	109.203.143.255	Mobile Communication Company of Iran Plc
109.203.144.0	109.203.159.255	Mobile Communication Company of Iran Plc
109.203.168.0	109.203.175.255	Farahoosh Dena PLC
109.203.176.0	109.203.183.255	Farahoosh Dena PLC
109.206.252.0	109.206.252.255	Afagh Andish Dadeh Pardis Co. Ltd
109.206.253.0	109.206.253.255	Asre Pardazeshe Ettelaate Amin Institute
109.206.254.0	109.206.254.255	Afagh Andish Dadeh Pardis Co. Ltd
109.206.255.0	109.206.255.255	Afagh Andish Dadeh Pardis Co. Ltd
109.225.128.0	109.225.159.255	Mobile Communication Company of Iran Plc
109.225.160.0	109.225.191.255	Mobile Communication Company of Iran Plc
109.230.64.0	109.230.71.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
109.230.72.0	109.230.79.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
109.230.80.0	109.230.80.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
109.230.81.0	109.230.81.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
109.230.82.0	109.230.83.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
109.230.84.0	109.230.87.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
109.230.88.0	109.230.95.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
109.230.192.0	109.230.193.255	Parsan Lin Co. PJS
109.230.200.0	109.230.200.255	Parsan Lin Co. PJS
109.230.204.0	109.230.204.255	Parsan Lin Co. PJS
109.230.205.0	109.230.205.255	Parsan Lin Co. PJS
109.230.206.0	109.230.206.255	Parsan Lin Co. PJS
109.230.207.0	109.230.207.255	Parsan Lin Co. PJS
109.230.221.0	109.230.221.255	Parsan Lin Co. PJS
109.230.223.0	109.230.223.255	Parsan Lin Co. PJS
109.230.242.0	109.230.242.255	Parsan Lin Co. PJS
109.230.246.0	109.230.246.255	Parsan Lin Co. PJS
109.230.247.0	109.230.247.255	Parsan Lin Co. PJS
109.230.251.0	109.230.251.255	Parsan Lin Co. PJS
109.232.0.0	109.232.3.255	Imen Sanat Novin Alighapoo Private Joint Stock Company
109.232.4.0	109.232.4.255	Imen Sanat Novin Alighapoo Private Joint Stock Company
109.232.5.0	109.232.5.255	Bank Refah Kargaran
109.232.6.0	109.232.6.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
109.232.7.0	109.232.7.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
109.238.176.0	109.238.191.255	Khalij Fars Ettela Resan LTD
109.239.0.0	109.239.1.255	Mehvar Machine
109.239.2.0	109.239.2.255	Mehvar Machine
109.239.3.0	109.239.3.255	Mehvar Machine
109.239.4.0	109.239.4.255	Mehvar Machine
109.239.5.0	109.239.5.255	Mehvar Machine
109.239.6.0	109.239.6.255	Mehvar Machine
109.239.7.0	109.239.7.255	Mehvar Machine
109.239.8.0	109.239.8.255	Mehvar Machine
109.239.9.0	109.239.9.255	Mehvar Machine
109.239.10.0	109.239.10.255	Mehvar Machine
109.239.11.0	109.239.11.255	Mehvar Machine
109.239.12.0	109.239.12.255	Mehvar Machine
109.239.13.0	109.239.13.255	Mehvar Machine
109.239.14.0	109.239.14.255	Mehvar Machine
109.239.15.0	109.239.15.255	Mehvar Machine
113.203.0.0	113.203.0.255	Mobile Communication Company of Iran Plc
113.203.1.0	113.203.1.255	Mobile Communication Company of Iran Plc
113.203.2.0	113.203.3.255	Mobile Communication Company of Iran Plc
113.203.4.0	113.203.7.255	Mobile Communication Company of Iran Plc
113.203.8.0	113.203.15.255	Mobile Communication Company of Iran Plc
113.203.16.0	113.203.31.255	Mobile Communication Company of Iran Plc
113.203.32.0	113.203.47.255	Mobile Communication Company of Iran Plc
113.203.48.0	113.203.63.255	Mobile Communication Company of Iran Plc
113.203.64.0	113.203.79.255	Mobile Communication Company of Iran Plc
113.203.80.0	113.203.95.255	Mobile Communication Company of Iran Plc
113.203.96.0	113.203.111.255	Mobile Communication Company of Iran Plc
113.203.112.0	113.203.127.255	Mobile Communication Company of Iran Plc
128.0.105.0	128.0.105.255	Noyan Abr Arvan Co. ( Private Joint Stock)
128.65.176.0	128.65.183.255	Asiatech Data Transmission Company
128.65.184.0	128.65.191.255	Asiatech Data Transmission Company
130.185.72.0	130.185.72.255	Vandad Vira Hooman LLC
130.185.73.0	130.185.73.255	Vandad Vira Hooman LLC
130.185.74.0	130.185.74.255	Vandad Vira Hooman LLC
130.185.75.0	130.185.75.255	Vandad Vira Hooman LLC
130.185.76.0	130.185.76.255	Vandad Vira Hooman LLC
130.185.77.0	130.185.77.255	Vandad Vira Hooman LLC
130.185.78.0	130.185.78.255	Vandad Vira Hooman LLC
130.185.79.0	130.185.79.255	Vandad Vira Hooman LLC
130.193.24.0	130.193.24.255	Fanavaran Asak Mahan Asia Ltd.
130.193.77.0	130.193.77.255	Sefroyek Pardaz Engineering PJSC
130.255.192.0	130.255.255.255	Mobile Communication Company of Iran Plc
134.255.196.0	134.255.197.255	Parsan Lin Co. PJS
134.255.200.0	134.255.200.255	Parsan Lin Co. PJS
134.255.201.0	134.255.201.255	Parsan Lin Co. PJS
134.255.202.0	134.255.202.255	Parsan Lin Co. PJS
134.255.203.0	134.255.203.255	Parsan Lin Co. PJS
134.255.204.0	134.255.204.255	Parsan Lin Co. PJS
134.255.205.0	134.255.205.255	Parsan Lin Co. PJS
134.255.206.0	134.255.206.255	Namavaran E Asr E Novin Information Technology (Fanhab) Private Joint-Stock Co
134.255.207.0	134.255.207.255	Parsan Lin Co. PJS
134.255.245.0	134.255.245.255	Parsan Lin Co. PJS
134.255.246.0	134.255.246.255	Parsan Lin Co. PJS
134.255.248.0	134.255.248.255	Fanavari Hamrah Payda Co. ( Private J.S.)
134.255.249.0	134.255.249.255	Parsan Lin Co. PJS
146.19.104.0	146.19.104.255	Charisma Financial Information Processing PJSC
146.19.212.0	146.19.212.255	Vandad Vira Hooman LLC
146.19.217.0	146.19.217.255	Saba Abr Mizban LLC
146.66.128.0	146.66.131.255	Pars Car International Communication Age PJSC
146.66.132.0	146.66.133.255	Pars Car International Communication Age PJSC
146.66.134.0	146.66.134.255	Saman Bank PJSC
146.66.135.0	146.66.135.255	Saman Bank PJSC
148.253.227.0	148.253.227.255	Asiatech Data Transmission Company
151.232.0.0	151.232.7.255	Iran Telecommunication Company Pjs
151.232.8.0	151.232.8.255	Iran Telecommunication Company Pjs
151.232.9.0	151.232.9.255	Iran Telecommunication Company Pjs
151.232.10.0	151.232.10.255	Iran Telecommunication Company Pjs
151.232.11.0	151.232.11.255	Iran Telecommunication Company Pjs
151.232.12.0	151.232.15.255	Iran Telecommunication Company Pjs
151.232.16.0	151.232.31.255	Iran Telecommunication Company Pjs
151.232.32.0	151.232.63.255	Iran Telecommunication Company Pjs
151.232.64.0	151.232.67.255	Iran Telecommunication Company Pjs
151.232.68.0	151.232.69.255	Iran Telecommunication Company Pjs
151.232.70.0	151.232.71.255	Iran Telecommunication Company Pjs
151.232.72.0	151.232.73.255	Iran Telecommunication Company Pjs
151.232.74.0	151.232.75.255	Iran Telecommunication Company Pjs
151.232.76.0	151.232.76.255	Iran Telecommunication Company Pjs
151.232.77.0	151.232.77.255	Iran Telecommunication Company Pjs
151.232.78.0	151.232.79.255	Iran Telecommunication Company Pjs
151.232.80.0	151.232.95.255	Iran Telecommunication Company Pjs
151.232.96.0	151.232.99.255	Iran Telecommunication Company Pjs
151.232.100.0	151.232.103.255	Iran Telecommunication Company Pjs
151.232.104.0	151.232.105.255	Iran Telecommunication Company Pjs
151.232.106.0	151.232.107.255	Iran Telecommunication Company Pjs
151.232.108.0	151.232.109.255	Iran Telecommunication Company Pjs
151.232.110.0	151.232.111.255	Iran Telecommunication Company Pjs
151.232.112.0	151.232.112.255	Shiraz University
151.232.113.0	151.232.113.255	Shiraz University
151.232.114.0	151.232.114.255	Shiraz University
151.232.115.0	151.232.115.255	Shiraz University
151.232.116.0	151.232.116.255	Iran Telecommunication Company Pjs
151.232.117.0	151.232.117.255	Iran Telecommunication Company Pjs
151.232.118.0	151.232.119.255	Iran Telecommunication Company Pjs
151.232.120.0	151.232.123.255	Iran Telecommunication Company Pjs
151.232.124.0	151.232.125.255	Iran Telecommunication Company Pjs
151.232.126.0	151.232.127.255	Iran Telecommunication Company Pjs
151.232.128.0	151.232.255.255	Iran Telecommunication Company Pjs
151.233.0.0	151.233.31.255	Iran Telecommunication Company Pjs
151.233.32.0	151.233.47.255	Iran Telecommunication Company Pjs
151.233.48.0	151.233.55.255	Iran Telecommunication Company Pjs
151.233.56.0	151.233.59.255	Iran Telecommunication Company Pjs
151.233.60.0	151.233.63.255	Iran Telecommunication Company Pjs
151.233.64.0	151.233.127.255	Iran Telecommunication Company Pjs
151.233.128.0	151.233.255.255	Iran Telecommunication Company Pjs
151.234.0.0	151.234.3.255	Iran Telecommunication Company Pjs
151.234.4.0	151.234.5.255	Iran Telecommunication Company Pjs
151.234.6.0	151.234.6.255	Iran Telecommunication Company Pjs
151.234.7.0	151.234.7.255	Iran Telecommunication Company Pjs
151.234.8.0	151.234.15.255	Iran Telecommunication Company Pjs
151.234.16.0	151.234.19.255	Iran Telecommunication Company Pjs
151.234.20.0	151.234.23.255	Iran Telecommunication Company Pjs
151.234.24.0	151.234.31.255	Iran Telecommunication Company Pjs
151.234.32.0	151.234.47.255	Iran Telecommunication Company Pjs
151.234.48.0	151.234.55.255	Iran Telecommunication Company Pjs
151.234.56.0	151.234.59.255	Iran Telecommunication Company Pjs
151.234.60.0	151.234.63.255	Iran Telecommunication Company Pjs
151.234.64.0	151.234.67.255	Iran Telecommunication Company Pjs
151.234.68.0	151.234.69.255	Iran Telecommunication Company Pjs
151.234.70.0	151.234.71.255	Iran Telecommunication Company Pjs
151.234.72.0	151.234.79.255	Iran Telecommunication Company Pjs
151.234.80.0	151.234.83.255	Iran Telecommunication Company Pjs
151.234.84.0	151.234.85.255	Iran Telecommunication Company Pjs
151.234.86.0	151.234.86.255	Iran Telecommunication Company Pjs
151.234.87.0	151.234.87.255	Iran Telecommunication Company Pjs
151.234.88.0	151.234.89.255	Iran Telecommunication Company Pjs
151.234.90.0	151.234.91.255	Iran Telecommunication Company Pjs
151.234.92.0	151.234.95.255	Iran Telecommunication Company Pjs
151.234.96.0	151.234.99.255	Iran Telecommunication Company Pjs
151.234.100.0	151.234.103.255	Iran Telecommunication Company Pjs
151.234.104.0	151.234.111.255	Iran Telecommunication Company Pjs
151.234.112.0	151.234.119.255	Iran Telecommunication Company Pjs
151.234.120.0	151.234.127.255	Iran Telecommunication Company Pjs
151.234.128.0	151.234.135.255	Iran Telecommunication Company Pjs
151.234.136.0	151.234.143.255	Iran Telecommunication Company Pjs
151.234.144.0	151.234.145.255	Iran Telecommunication Company Pjs
151.234.146.0	151.234.147.255	Iran Telecommunication Company Pjs
151.234.148.0	151.234.148.255	Iran Telecommunication Company Pjs
151.234.149.0	151.234.149.255	Iran Telecommunication Company Pjs
151.234.150.0	151.234.151.255	Iran Telecommunication Company Pjs
151.234.152.0	151.234.153.255	Iran Telecommunication Company Pjs
151.234.154.0	151.234.155.255	Iran Telecommunication Company Pjs
151.234.156.0	151.234.159.255	Iran Telecommunication Company Pjs
151.234.160.0	151.234.160.255	Iran Telecommunication Company Pjs
151.234.161.0	151.234.161.255	Iran Telecommunication Company Pjs
151.234.162.0	151.234.162.255	Iran Telecommunication Company Pjs
151.234.163.0	151.234.163.255	Iran Telecommunication Company Pjs
151.234.164.0	151.234.164.255	Iran Telecommunication Company Pjs
151.234.165.0	151.234.165.255	Iran Telecommunication Company Pjs
151.234.166.0	151.234.166.255	Iran Telecommunication Company Pjs
151.234.167.0	151.234.167.255	Iran Telecommunication Company Pjs
151.234.168.0	151.234.168.255	Iran Telecommunication Company Pjs
151.234.169.0	151.234.169.255	Iran Telecommunication Company Pjs
151.234.170.0	151.234.171.255	Iran Telecommunication Company Pjs
151.234.172.0	151.234.173.255	Iran Telecommunication Company Pjs
151.234.174.0	151.234.174.255	Iran Telecommunication Company Pjs
151.234.175.0	151.234.175.255	Iran Telecommunication Company Pjs
151.234.176.0	151.234.183.255	Iran Telecommunication Company Pjs
151.234.184.0	151.234.185.255	Iran Telecommunication Company Pjs
151.234.186.0	151.234.187.255	Iran Telecommunication Company Pjs
151.234.188.0	151.234.191.255	Iran Telecommunication Company Pjs
151.234.192.0	151.234.192.255	Iran Telecommunication Company Pjs
151.234.193.0	151.234.193.255	Iran Telecommunication Company Pjs
151.234.194.0	151.234.195.255	Iran Telecommunication Company Pjs
151.234.196.0	151.234.197.255	Iran Telecommunication Company Pjs
151.234.198.0	151.234.199.255	Iran Telecommunication Company Pjs
151.234.200.0	151.234.201.255	Iran Telecommunication Company Pjs
151.234.202.0	151.234.203.255	Iran Telecommunication Company Pjs
151.234.204.0	151.234.207.255	Iran Telecommunication Company Pjs
151.234.208.0	151.234.223.255	Iran Telecommunication Company Pjs
151.234.224.0	151.234.239.255	Iran Telecommunication Company Pjs
151.234.240.0	151.234.240.255	Iran Telecommunication Company Pjs
151.234.241.0	151.234.241.255	Iran Telecommunication Company Pjs
151.234.242.0	151.234.242.255	Iran Telecommunication Company Pjs
151.234.243.0	151.234.243.255	Iran Telecommunication Company Pjs
151.234.244.0	151.234.245.255	Iran Telecommunication Company Pjs
151.234.246.0	151.234.247.255	Iran Telecommunication Company Pjs
151.234.248.0	151.234.251.255	Iran Telecommunication Company Pjs
151.234.252.0	151.234.255.255	Iran Telecommunication Company Pjs
151.235.0.0	151.235.31.255	Iran Telecommunication Company Pjs
151.235.32.0	151.235.63.255	Iran Telecommunication Company Pjs
151.235.64.0	151.235.127.255	Iran Telecommunication Company Pjs
151.235.128.0	151.235.143.255	Iran Telecommunication Company Pjs
151.235.144.0	151.235.159.255	Iran Telecommunication Company Pjs
151.235.160.0	151.235.175.255	Iran Telecommunication Company Pjs
151.235.176.0	151.235.179.255	Iran Telecommunication Company Pjs
151.235.180.0	151.235.183.255	Iran Telecommunication Company Pjs
151.235.184.0	151.235.191.255	Iran Telecommunication Company Pjs
151.235.192.0	151.235.255.255	Iran Telecommunication Company Pjs
151.238.0.0	151.238.15.255	Aria Shatel PJSC
151.238.16.0	151.238.31.255	Aria Shatel PJSC
151.238.32.0	151.238.47.255	Aria Shatel PJSC
151.238.48.0	151.238.63.255	Aria Shatel PJSC
151.238.64.0	151.238.95.255	Aria Shatel PJSC
151.238.96.0	151.238.99.255	Aria Shatel PJSC
151.238.100.0	151.238.103.255	Aria Shatel PJSC
151.238.104.0	151.238.107.255	Aria Shatel PJSC
151.238.108.0	151.238.111.255	Aria Shatel PJSC
151.238.112.0	151.238.115.255	Aria Shatel PJSC
151.238.116.0	151.238.119.255	Aria Shatel PJSC
151.238.120.0	151.238.123.255	Aria Shatel PJSC
151.238.124.0	151.238.127.255	Aria Shatel PJSC
151.238.128.0	151.238.131.255	Aria Shatel PJSC
151.238.132.0	151.238.135.255	Aria Shatel PJSC
151.238.136.0	151.238.139.255	Aria Shatel PJSC
151.238.140.0	151.238.143.255	Aria Shatel PJSC
151.238.144.0	151.238.147.255	Aria Shatel PJSC
151.238.148.0	151.238.151.255	Aria Shatel PJSC
151.238.152.0	151.238.155.255	Aria Shatel PJSC
151.238.156.0	151.238.159.255	Aria Shatel PJSC
151.238.160.0	151.238.163.255	Aria Shatel PJSC
151.238.164.0	151.238.167.255	Aria Shatel PJSC
151.238.168.0	151.238.171.255	Aria Shatel PJSC
151.238.172.0	151.238.175.255	Aria Shatel PJSC
151.238.176.0	151.238.179.255	Aria Shatel PJSC
151.238.180.0	151.238.183.255	Aria Shatel PJSC
151.238.184.0	151.238.191.255	Aria Shatel PJSC
151.238.192.0	151.238.255.255	Aria Shatel PJSC
151.239.0.0	151.239.63.255	Aria Shatel PJSC
151.239.64.0	151.239.127.255	Aria Shatel PJSC
151.239.128.0	151.239.191.255	Aria Shatel PJSC
151.239.192.0	151.239.255.255	Aria Shatel PJSC
152.89.12.0	152.89.15.255	Sharif University of Technology
152.89.44.0	152.89.44.255	Sefroyek Pardaz Engineering PJSC
152.89.45.0	152.89.45.255	Sefroyek Pardaz Engineering PJSC
152.89.46.0	152.89.46.255	Sefroyek Pardaz Engineering PJSC
152.89.47.0	152.89.47.255	Sefroyek Pardaz Engineering PJSC
153.51.0.0	153.51.31.255	Mobin Net Communication Company (Private Joint Stock)
153.51.128.0	153.51.159.255	Mobin Net Communication Company (Private Joint Stock)
157.119.188.0	157.119.188.255	Sindad Network Technology PJSC
157.119.189.0	157.119.189.255	Charisma Financial Information Processing PJSC
157.119.190.0	157.119.190.255	Sindad Network Technology PJSC
157.119.191.0	157.119.191.255	Sindad Network Technology PJSC
158.58.0.0	158.58.15.255	Mobile Communication Company of Iran Plc
158.58.16.0	158.58.31.255	Mobile Communication Company of Iran Plc
158.58.32.0	158.58.47.255	Mobile Communication Company of Iran Plc
158.58.48.0	158.58.63.255	Mobile Communication Company of Iran Plc
158.58.64.0	158.58.79.255	Mobile Communication Company of Iran Plc
158.58.80.0	158.58.95.255	Mobile Communication Company of Iran Plc
158.58.96.0	158.58.111.255	Mobile Communication Company of Iran Plc
158.58.112.0	158.58.127.255	Mobile Communication Company of Iran Plc
158.58.184.0	158.58.184.255	Faraso Samaneh Pasargad Ltd.
158.58.185.0	158.58.185.255	Faraso Samaneh Pasargad Ltd.
158.58.186.0	158.58.186.255	Faraso Samaneh Pasargad Ltd.
158.58.187.0	158.58.187.255	Faraso Samaneh Pasargad Ltd.
158.58.188.0	158.58.188.255	Faraso Samaneh Pasargad Ltd.
158.58.189.0	158.58.189.255	Faraso Samaneh Pasargad Ltd.
158.58.190.0	158.58.190.255	Faraso Samaneh Pasargad Ltd.
158.58.191.0	158.58.191.255	Faraso Samaneh Pasargad Ltd.
158.255.74.0	158.255.74.255	Vandad Vira Hooman LLC
158.255.78.0	158.255.78.255	Sefroyek Pardaz Engineering PJSC
159.20.96.0	159.20.111.255	Andishe Sabz Khazar Co. P.J.S.
164.40.233.0	164.40.233.255	Iran Chamber of Commerce Industries Mines and Agriculture
164.138.16.0	164.138.16.255	Ravand Tazeh Co .Pjs.
164.138.17.0	164.138.17.255	Ravand Tazeh Co .Pjs.
164.138.18.0	164.138.18.255	Ravand Tazeh Co .Pjs.
164.138.19.0	164.138.19.255	Ravand Tazeh Co .Pjs.
164.138.20.0	164.138.20.255	Ravand Tazeh Co .Pjs.
164.138.21.0	164.138.21.255	Ravand Tazeh Co .Pjs.
164.138.22.0	164.138.22.255	Ravand Tazeh Co .Pjs.
164.138.23.0	164.138.23.255	Ravand Tazeh Co .Pjs.
164.138.128.0	164.138.129.255	Mobile Communication Company of Iran Plc
164.138.130.0	164.138.130.255	Mobile Communication Company of Iran Plc
164.138.131.0	164.138.131.255	Mobile Communication Company of Iran Plc
164.138.132.0	164.138.135.255	Mobile Communication Company of Iran Plc
164.138.136.0	164.138.143.255	Mobile Communication Company of Iran Plc
164.138.144.0	164.138.159.255	Mobile Communication Company of Iran Plc
164.138.160.0	164.138.175.255	Mobile Communication Company of Iran Plc
164.138.176.0	164.138.191.255	Mobile Communication Company of Iran Plc
164.138.203.0	164.138.203.255	Pooya Parto Qeshm Cooperative Company
164.138.204.0	164.138.204.255	Khalij Fars Ettela Resan LTD
164.138.206.0	164.138.206.255	Ideh Gozin Ertebatat E Romak Co. Ltd
164.215.56.0	164.215.57.255	Agriculture Bank PJSC
164.215.58.0	164.215.59.255	Agriculture Bank PJSC
164.215.60.0	164.215.60.255	Agriculture Bank PJSC
164.215.61.0	164.215.61.255	Agriculture Bank PJSC
164.215.62.0	164.215.62.255	Agriculture Bank PJSC
164.215.63.0	164.215.63.255	Agriculture Bank PJSC
164.215.128.0	164.215.159.255	Asiatech Data Transmission Company
164.215.160.0	164.215.175.255	Asiatech Data Transmission Company
164.215.176.0	164.215.191.255	Asiatech Data Transmission Company
164.215.192.0	164.215.195.255	Asiatech Data Transmission Company
164.215.196.0	164.215.199.255	Asiatech Data Transmission Company
164.215.200.0	164.215.203.255	Asiatech Data Transmission Company
164.215.204.0	164.215.207.255	Asiatech Data Transmission Company
164.215.208.0	164.215.211.255	Asiatech Data Transmission Company
164.215.212.0	164.215.215.255	Asiatech Data Transmission Company
164.215.216.0	164.215.219.255	Asiatech Data Transmission Company
164.215.220.0	164.215.223.255	Asiatech Data Transmission Company
164.215.224.0	164.215.227.255	Asiatech Data Transmission Company
164.215.228.0	164.215.231.255	Asiatech Data Transmission Company
164.215.232.0	164.215.235.255	Asiatech Data Transmission Company
164.215.236.0	164.215.239.255	Asiatech Data Transmission Company
164.215.240.0	164.215.243.255	Asiatech Data Transmission Company
164.215.244.0	164.215.247.255	Asiatech Data Transmission Company
164.215.248.0	164.215.251.255	Asiatech Data Transmission Company
164.215.252.0	164.215.255.255	Asiatech Data Transmission Company
171.22.24.0	171.22.24.255	Vandad Vira Hooman LLC
171.22.25.0	171.22.25.255	Vandad Vira Hooman LLC
171.22.26.0	171.22.26.255	Vandad Vira Hooman LLC
171.22.27.0	171.22.27.255	Vandad Vira Hooman LLC
172.80.128.0	172.80.128.255	Mobile Communication Company of Iran Plc
172.80.129.0	172.80.129.255	Mobile Communication Company of Iran Plc
172.80.130.0	172.80.131.255	Mobile Communication Company of Iran Plc
172.80.132.0	172.80.135.255	Mobile Communication Company of Iran Plc
172.80.136.0	172.80.143.255	Mobile Communication Company of Iran Plc
172.80.144.0	172.80.159.255	Mobile Communication Company of Iran Plc
172.80.160.0	172.80.175.255	Mobile Communication Company of Iran Plc
172.80.176.0	172.80.191.255	Mobile Communication Company of Iran Plc
172.80.192.0	172.80.192.255	Mobile Communication Company of Iran Plc
172.80.193.0	172.80.193.255	Mobile Communication Company of Iran Plc
172.80.194.0	172.80.195.255	Mobile Communication Company of Iran Plc
172.80.196.0	172.80.199.255	Mobile Communication Company of Iran Plc
172.80.200.0	172.80.207.255	Mobile Communication Company of Iran Plc
172.80.208.0	172.80.223.255	Mobile Communication Company of Iran Plc
172.80.224.0	172.80.239.255	Mobile Communication Company of Iran Plc
172.80.240.0	172.80.247.255	Mobile Communication Company of Iran Plc
172.80.248.0	172.80.251.255	Mobile Communication Company of Iran Plc
172.80.252.0	172.80.252.255	Mobile Communication Company of Iran Plc
172.80.253.0	172.80.253.255	Mobile Communication Company of Iran Plc
172.80.254.0	172.80.254.255	Mobile Communication Company of Iran Plc
172.80.255.0	172.80.255.255	Mobile Communication Company of Iran Plc
176.10.95.0	176.10.95.255	Cooperative Afra ertebatat-e-sabet-e Rasa Co
176.46.134.0	176.46.134.255	Farahoosh Dena PLC
176.46.138.0	176.46.139.255	Farahoosh Dena PLC
176.46.141.0	176.46.141.255	Farahoosh Dena PLC
176.46.144.0	176.46.144.255	Farahoosh Dena PLC
176.46.147.0	176.46.147.255	Farahoosh Dena PLC
176.46.148.0	176.46.149.255	Farahoosh Dena PLC
176.46.150.0	176.46.150.255	Farahoosh Dena PLC
176.46.152.0	176.46.155.255	Farahoosh Dena PLC
176.46.156.0	176.46.157.255	Farahoosh Dena PLC
176.46.158.0	176.46.158.255	Farahoosh Dena PLC
176.56.144.0	176.56.144.255	Iranian Information Technology PJSC
176.56.145.0	176.56.145.255	Iranian Information Technology Company PLC
176.56.146.0	176.56.146.255	Iranian Information Technology PJSC
176.56.147.0	176.56.147.255	Iranian Information Technology Company PLC
176.56.148.0	176.56.148.255	Iranian Information Technology PJSC
176.56.149.0	176.56.149.255	Iranian Information Technology Company PLC
176.56.150.0	176.56.150.255	Iranian Information Technology PJSC
176.56.151.0	176.56.151.255	Iranian Information Technology Company PLC
176.56.152.0	176.56.155.255	Iranian Information Technology Company PLC
176.56.156.0	176.56.156.255	Secure Infrastructure of Transactional Services Company PJSC
176.56.157.0	176.56.157.255	Secure Infrastructure of Transactional Services Company PJSC
176.56.158.0	176.56.158.255	Secure Infrastructure of Transactional Services Company PJSC
176.56.159.0	176.56.159.255	Secure Infrastructure of Transactional Services Company PJSC
176.62.144.0	176.62.151.255	KEYANA Information Technology Co. Ltd.
176.65.133.0	176.65.133.255	Abr Afzar Sedna Ertebat Company (LTD.)
176.65.160.0	176.65.191.255	Iran Telecommunication Company Pjs
176.65.192.0	176.65.192.255	Mobile Communication Company of Iran Plc
176.65.193.0	176.65.193.255	Mobile Communication Company of Iran Plc
176.65.194.0	176.65.195.255	Mobile Communication Company of Iran Plc
176.65.196.0	176.65.199.255	Mobile Communication Company of Iran Plc
176.65.200.0	176.65.207.255	Mobile Communication Company of Iran Plc
176.65.208.0	176.65.223.255	Mobile Communication Company of Iran Plc
176.65.224.0	176.65.227.255	Iran Telecommunication Company Pjs
176.65.228.0	176.65.231.255	Iran Telecommunication Company Pjs
176.65.232.0	176.65.239.255	Iran Telecommunication Company Pjs
176.65.240.0	176.65.240.255	Asiatech Data Transmission Company
176.65.241.0	176.65.241.255	Asiatech Data Transmission Company
176.65.242.0	176.65.242.255	Asiatech Data Transmission Company
176.65.243.0	176.65.243.255	Asiatech Data Transmission Company
176.65.244.0	176.65.247.255	Iran Telecommunication Company Pjs
176.65.248.0	176.65.251.255	Iran Telecommunication Company Pjs
176.65.252.0	176.65.253.255	Asiatech Data Transmission Company
176.65.254.0	176.65.254.255	Asiatech Data Transmission Company
176.65.255.0	176.65.255.255	Asiatech Data Transmission Company
176.67.64.0	176.67.79.255	Khalij Fars Ettela Resan LTD
176.97.218.0	176.97.218.255	Vandad Vira Hooman LLC
176.97.220.0	176.97.220.255	Asre Pardazeshe Ettelaate Amin Institute
176.101.32.0	176.101.32.255	Respina Networks & Beyond PJSC
176.101.33.0	176.101.33.255	Respina Networks & Beyond PJSC
176.101.48.0	176.101.48.255	Isfahan University of Technology
176.101.49.0	176.101.49.255	Isfahan University of Technology
176.101.50.0	176.101.50.255	Isfahan University of Technology
176.101.51.0	176.101.51.255	Isfahan University of Technology
176.101.52.0	176.101.52.255	Isfahan University of Technology
176.101.53.0	176.101.53.255	Isfahan University of Technology
176.101.54.0	176.101.54.255	Isfahan University of Technology
176.101.55.0	176.101.55.255	Isfahan University of Technology
176.102.224.0	176.102.239.255	Isfahan University of Medical Sciences & Health Services
176.102.240.0	176.102.255.255	Isfahan University of Medical Sciences & Health Services
176.105.245.0	176.105.245.255	Ertebatat Shabakeh Risman Ltd
176.116.7.0	176.116.7.255	Khatam Academic Institute
176.117.107.0	176.117.107.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
176.120.16.0	176.120.16.255	Vandad Vira Hooman LLC
176.120.17.0	176.120.17.255	Ali Monfared
176.120.19.0	176.120.19.255	Web Dadeh Paydar Co (Ltd)
176.122.210.0	176.122.210.255	Gostaresh Dade Ako Fanavaran Co LLC
176.122.211.0	176.122.211.255	Gostaresh Dade Ako Fanavaran Co LLC
176.123.64.0	176.123.64.255	Shiraz University of Medical Science and Health Care Services
176.123.65.0	176.123.65.255	Shiraz University of Medical Science and Health Care Services
176.123.66.0	176.123.66.255	Shiraz University of Medical Science and Health Care Services
176.123.67.0	176.123.67.255	Shiraz University of Medical Science and Health Care Services
176.123.68.0	176.123.68.255	Shiraz University of Medical Science and Health Care Services
176.123.69.0	176.123.69.255	Shiraz University of Medical Science and Health Care Services
176.123.70.0	176.123.70.255	Shiraz University of Medical Science and Health Care Services
176.123.71.0	176.123.71.255	Shiraz University of Medical Science and Health Care Services
176.123.72.0	176.123.79.255	Shiraz University of Medical Science and Health Care Services
176.123.80.0	176.123.95.255	Shiraz University of Medical Science and Health Care Services
176.123.96.0	176.123.96.255	Shiraz University of Medical Science and Health Care Services
176.123.97.0	176.123.97.255	Shiraz University of Medical Science and Health Care Services
176.123.98.0	176.123.98.255	Shiraz University of Medical Science and Health Care Services
176.123.99.0	176.123.99.255	Shiraz University of Medical Science and Health Care Services
176.123.100.0	176.123.100.255	Shiraz University of Medical Science and Health Care Services
176.123.101.0	176.123.101.255	Shiraz University of Medical Science and Health Care Services
176.123.102.0	176.123.102.255	Shiraz University of Medical Science and Health Care Services
176.123.103.0	176.123.103.255	Shiraz University of Medical Science and Health Care Services
176.123.104.0	176.123.111.255	Shiraz University of Medical Science and Health Care Services
176.123.112.0	176.123.112.255	Shiraz University of Medical Science and Health Care Services
176.123.113.0	176.123.113.255	Shiraz University of Medical Science and Health Care Services
176.123.114.0	176.123.114.255	Shiraz University of Medical Science and Health Care Services
176.123.115.0	176.123.115.255	Shiraz University of Medical Science and Health Care Services
176.123.116.0	176.123.116.255	Shiraz University of Medical Science and Health Care Services
176.123.117.0	176.123.117.255	Shiraz University of Medical Science and Health Care Services
176.123.118.0	176.123.119.255	Shiraz University of Medical Science and Health Care Services
176.123.120.0	176.123.127.255	Shiraz University of Medical Science and Health Care Services
176.126.120.0	176.126.120.255	Tida Data Processing Ltd
176.126.223.0	176.126.223.255	Mobin Net Communication Company (Private Joint Stock)
176.221.64.0	176.221.64.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
176.221.65.0	176.221.65.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
176.221.66.0	176.221.67.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
176.221.68.0	176.221.69.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
176.221.70.0	176.221.70.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
176.221.71.0	176.221.71.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
176.223.80.0	176.223.87.255	Iran Telecommunication Company Pjs
178.21.40.0	178.21.40.255	Parsian Bank (Public Joint-Stock)
178.21.41.0	178.21.41.255	Parsian Bank (Public Joint-Stock)
178.21.42.0	178.21.42.255	Parsian Bank (Public Joint-Stock)
178.21.43.0	178.21.43.255	Caspian Applied Systems Services Supplying Company (Private Joint-Stock)
178.21.44.0	178.21.44.255	Caspian Applied Systems Services Supplying Company (Private Joint-Stock)
178.21.45.0	178.21.45.255	Caspian Applied Systems Services Supplying Company (Private Joint-Stock)
178.21.46.0	178.21.46.255	Caspian Applied Systems Services Supplying Company (Private Joint-Stock)
178.21.47.0	178.21.47.255	Caspian Applied Systems Services Supplying Company (Private Joint-Stock)
178.21.160.0	178.21.161.255	Pishtaze e Jadid Information Company (Ltd.)
178.21.162.0	178.21.162.255	Pishtaze e Jadid Information Company (Ltd.)
178.21.163.0	178.21.163.255	Pishtaze e Jadid Information Company (Ltd.)
178.21.164.0	178.21.164.255	Pishtaze e Jadid Information Company (Ltd.)
178.21.165.0	178.21.165.255	Pishtaze e Jadid Information Company (Ltd.)
178.21.166.0	178.21.167.255	Pishtaze e Jadid Information Company (Ltd.)
178.22.72.0	178.22.79.255	Fars News Agency Cultural Arts Institute
178.22.120.0	178.22.120.255	Asiatech Data Transmission Company
178.22.121.0	178.22.121.255	Asiatech Data Transmission Company
178.22.122.0	178.22.122.255	Asiatech Data Transmission Company
178.22.123.0	178.22.123.255	Asiatech Data Transmission Company
178.22.124.0	178.22.124.255	Asiatech Data Transmission Company
178.22.125.0	178.22.125.255	Asiatech Data Transmission Company
178.22.126.0	178.22.126.255	Asiatech Data Transmission Company
178.22.127.0	178.22.127.255	Asiatech Data Transmission Company
178.131.0.0	178.131.3.255	Mobin Net Communication Company (Private Joint Stock)
178.131.4.0	178.131.4.255	Mobin Net Communication Company (Private Joint Stock)
178.131.5.0	178.131.5.255	Mobin Net Communication Company (Private Joint Stock)
178.131.6.0	178.131.7.255	Mobin Net Communication Company (Private Joint Stock)
178.131.8.0	178.131.15.255	Mobin Net Communication Company (Private Joint Stock)
178.131.16.0	178.131.19.255	Mobin Net Communication Company (Private Joint Stock)
178.131.20.0	178.131.23.255	Mobin Net Communication Company (Private Joint Stock)
178.131.24.0	178.131.24.255	Mobin Net Communication Company (Private Joint Stock)
178.131.25.0	178.131.25.255	Mobin Net Communication Company (Private Joint Stock)
178.131.26.0	178.131.26.255	Mobin Net Communication Company (Private Joint Stock)
178.131.27.0	178.131.27.255	Mobin Net Communication Company (Private Joint Stock)
178.131.28.0	178.131.31.255	Mobin Net Communication Company (Private Joint Stock)
178.131.32.0	178.131.39.255	Mobin Net Communication Company (Private Joint Stock)
178.131.40.0	178.131.47.255	Mobin Net Communication Company (Private Joint Stock)
178.131.48.0	178.131.63.255	Mobin Net Communication Company (Private Joint Stock)
178.131.64.0	178.131.79.255	Mobin Net Communication Company (Private Joint Stock)
178.131.80.0	178.131.95.255	Mobin Net Communication Company (Private Joint Stock)
178.131.96.0	178.131.99.255	Mobin Net Communication Company (Private Joint Stock)
178.131.100.0	178.131.103.255	Mobin Net Communication Company (Private Joint Stock)
178.131.104.0	178.131.111.255	Mobin Net Communication Company (Private Joint Stock)
178.131.112.0	178.131.115.255	Mobin Net Communication Company (Private Joint Stock)
178.131.116.0	178.131.119.255	Mobin Net Communication Company (Private Joint Stock)
178.131.120.0	178.131.120.255	Mobile Communication Company of Iran Plc
178.131.121.0	178.131.121.255	Mobile Communication Company of Iran Plc
178.131.122.0	178.131.122.255	Mobile Communication Company of Iran Plc
178.131.123.0	178.131.123.255	Mobile Communication Company of Iran Plc
178.131.124.0	178.131.124.255	Mobile Communication Company of Iran Plc
178.131.125.0	178.131.125.255	Mobile Communication Company of Iran Plc
178.131.126.0	178.131.126.255	Mobile Communication Company of Iran Plc
178.131.127.0	178.131.127.255	Mobile Communication Company of Iran Plc
178.131.128.0	178.131.128.255	Mobin Net Communication Company (Private Joint Stock)
178.131.129.0	178.131.129.255	Mobin Net Communication Company (Private Joint Stock)
178.131.130.0	178.131.130.255	Mobin Net Communication Company (Private Joint Stock)
178.131.131.0	178.131.131.255	Mobin Net Communication Company (Private Joint Stock)
178.131.132.0	178.131.135.255	Mobin Net Communication Company (Private Joint Stock)
178.131.136.0	178.131.136.255	Mobin Net Communication Company (Private Joint Stock)
178.131.137.0	178.131.137.255	Mobin Net Communication Company (Private Joint Stock)
178.131.138.0	178.131.138.255	Mobin Net Communication Company (Private Joint Stock)
178.131.139.0	178.131.139.255	Mobin Net Communication Company (Private Joint Stock)
178.131.140.0	178.131.143.255	Mobin Net Communication Company (Private Joint Stock)
178.131.144.0	178.131.144.255	Mobin Net Communication Company (Private Joint Stock)
178.131.145.0	178.131.145.255	Mobin Net Communication Company (Private Joint Stock)
178.131.146.0	178.131.146.255	Mobin Net Communication Company (Private Joint Stock)
178.131.147.0	178.131.147.255	Mobin Net Communication Company (Private Joint Stock)
178.131.148.0	178.131.148.255	Mobin Net Communication Company (Private Joint Stock)
178.131.149.0	178.131.149.255	Mobin Net Communication Company (Private Joint Stock)
178.131.150.0	178.131.151.255	Mobin Net Communication Company (Private Joint Stock)
178.131.152.0	178.131.152.255	Mobin Net Communication Company (Private Joint Stock)
178.131.153.0	178.131.153.255	Mobin Net Communication Company (Private Joint Stock)
178.131.154.0	178.131.154.255	Mobin Net Communication Company (Private Joint Stock)
178.131.155.0	178.131.155.255	Mobin Net Communication Company (Private Joint Stock)
178.131.156.0	178.131.159.255	Mobin Net Communication Company (Private Joint Stock)
178.131.160.0	178.131.160.255	Mobin Net Communication Company (Private Joint Stock)
178.131.161.0	178.131.161.255	Mobin Net Communication Company (Private Joint Stock)
178.131.162.0	178.131.163.255	Mobin Net Communication Company (Private Joint Stock)
178.131.164.0	178.131.167.255	Mobin Net Communication Company (Private Joint Stock)
178.131.168.0	178.131.169.255	Mobin Net Communication Company (Private Joint Stock)
178.131.170.0	178.131.170.255	Mobin Net Communication Company (Private Joint Stock)
178.131.171.0	178.131.171.255	Mobin Net Communication Company (Private Joint Stock)
178.131.172.0	178.131.175.255	Mobin Net Communication Company (Private Joint Stock)
178.131.176.0	178.131.176.255	Mobin Net Communication Company (Private Joint Stock)
178.131.177.0	178.131.177.255	Mobin Net Communication Company (Private Joint Stock)
178.131.178.0	178.131.178.255	Mobin Net Communication Company (Private Joint Stock)
178.131.179.0	178.131.179.255	Mobin Net Communication Company (Private Joint Stock)
178.131.180.0	178.131.183.255	Mobin Net Communication Company (Private Joint Stock)
178.131.184.0	178.131.184.255	Mobin Net Communication Company (Private Joint Stock)
178.131.185.0	178.131.185.255	Mobin Net Communication Company (Private Joint Stock)
178.131.186.0	178.131.186.255	Mobin Net Communication Company (Private Joint Stock)
178.131.187.0	178.131.187.255	Mobin Net Communication Company (Private Joint Stock)
178.131.188.0	178.131.189.255	Mobin Net Communication Company (Private Joint Stock)
178.131.190.0	178.131.191.255	Mobin Net Communication Company (Private Joint Stock)
178.131.192.0	178.131.192.255	Mobile Communication Company of Iran Plc
178.131.193.0	178.131.193.255	Mobile Communication Company of Iran Plc
178.131.194.0	178.131.195.255	Mobile Communication Company of Iran Plc
178.131.196.0	178.131.199.255	Mobile Communication Company of Iran Plc
178.131.200.0	178.131.207.255	Mobile Communication Company of Iran Plc
178.131.208.0	178.131.223.255	Mobile Communication Company of Iran Plc
178.131.224.0	178.131.239.255	Mobile Communication Company of Iran Plc
178.131.240.0	178.131.255.255	Mobile Communication Company of Iran Plc
178.157.1.0	178.157.1.255	Shabakieh Isfahan Co PJSC
178.173.128.0	178.173.128.255	Shiraz Hamyar Co.
178.173.129.0	178.173.129.255	Shiraz Hamyar Co.
178.173.130.0	178.173.130.255	Shiraz Hamyar Co.
178.173.131.0	178.173.131.255	Shiraz Hamyar Co.
178.173.132.0	178.173.132.255	Shiraz Hamyar Co.
178.173.133.0	178.173.133.255	Shiraz Hamyar Co.
178.173.134.0	178.173.134.255	Shiraz Hamyar Co.
178.173.135.0	178.173.135.255	Shiraz Hamyar Co.
178.173.136.0	178.173.136.255	Shiraz Hamyar Co.
178.173.137.0	178.173.137.255	Shiraz Hamyar Co.
178.173.138.0	178.173.138.255	Shiraz Hamyar Co.
178.173.139.0	178.173.139.255	Shiraz Hamyar Co.
178.173.140.0	178.173.140.255	Shiraz Hamyar Co.
178.173.141.0	178.173.141.255	Shiraz Hamyar Co.
178.173.142.0	178.173.142.255	Shiraz Hamyar Co.
178.173.143.0	178.173.143.255	Shiraz Hamyar Co.
178.173.144.0	178.173.144.255	Shiraz Hamyar Co.
178.173.145.0	178.173.145.255	Shiraz Hamyar Co.
178.173.146.0	178.173.146.255	Shiraz Hamyar Co.
178.173.147.0	178.173.147.255	Shiraz Hamyar Co.
178.173.148.0	178.173.148.255	Shiraz Hamyar Co.
178.173.149.0	178.173.149.255	Shiraz Hamyar Co.
178.173.150.0	178.173.150.255	Shiraz Hamyar Co.
178.173.151.0	178.173.151.255	Shiraz Hamyar Co.
178.173.152.0	178.173.152.255	Shiraz University
178.173.153.0	178.173.153.255	Shiraz University
178.173.154.0	178.173.154.255	Shiraz University
178.173.155.0	178.173.155.255	Shiraz University
178.173.156.0	178.173.156.255	Shiraz University
178.173.157.0	178.173.157.255	Shiraz University
178.173.158.0	178.173.158.255	Shiraz University
178.173.159.0	178.173.159.255	Shiraz University
178.173.160.0	178.173.160.255	Shiraz Hamyar Co.
178.173.161.0	178.173.161.255	Shiraz Hamyar Co.
178.173.162.0	178.173.162.255	Shiraz Hamyar Co.
178.173.163.0	178.173.163.255	Shiraz Hamyar Co.
178.173.164.0	178.173.164.255	Shiraz Hamyar Co.
178.173.165.0	178.173.165.255	Shiraz Hamyar Co.
178.173.166.0	178.173.166.255	Shiraz Hamyar Co.
178.173.167.0	178.173.167.255	Shiraz Hamyar Co.
178.173.168.0	178.173.168.255	Shiraz Hamyar Co.
178.173.169.0	178.173.169.255	Shiraz Hamyar Co.
178.173.170.0	178.173.170.255	Shiraz Hamyar Co.
178.173.171.0	178.173.171.255	Shiraz Hamyar Co.
178.173.172.0	178.173.172.255	Shiraz Hamyar Co.
178.173.173.0	178.173.173.255	Shiraz Hamyar Co.
178.173.174.0	178.173.174.255	Shiraz Hamyar Co.
178.173.175.0	178.173.175.255	Shiraz Hamyar Co.
178.173.176.0	178.173.176.255	Shiraz Hamyar Co.
178.173.177.0	178.173.177.255	Shiraz Hamyar Co.
178.173.178.0	178.173.178.255	Shiraz Hamyar Co.
178.173.179.0	178.173.179.255	Shiraz Hamyar Co.
178.173.180.0	178.173.180.255	Shiraz Hamyar Co.
178.173.181.0	178.173.181.255	Shiraz Hamyar Co.
178.173.182.0	178.173.182.255	Shiraz Hamyar Co.
178.173.183.0	178.173.183.255	Shiraz Hamyar Co.
178.173.184.0	178.173.184.255	Shiraz Hamyar Co.
178.173.185.0	178.173.185.255	Shiraz Hamyar Co.
178.173.186.0	178.173.186.255	Shiraz Hamyar Co.
178.173.187.0	178.173.187.255	Shiraz Hamyar Co.
178.173.188.0	178.173.188.255	Shiraz Hamyar Co.
178.173.189.0	178.173.189.255	Shiraz Hamyar Co.
178.173.190.0	178.173.190.255	Shiraz Hamyar Co.
178.173.191.0	178.173.191.255	Shiraz Hamyar Co.
178.173.192.0	178.173.192.255	Shiraz Hamyar Co.
178.173.193.0	178.173.193.255	Shiraz Hamyar Co.
178.173.194.0	178.173.194.255	Shiraz Hamyar Co.
178.173.195.0	178.173.195.255	Shiraz Hamyar Co.
178.173.196.0	178.173.196.255	Shiraz Hamyar Co.
178.173.197.0	178.173.197.255	Shiraz Hamyar Co.
178.173.198.0	178.173.198.255	Shiraz Hamyar Co.
178.173.199.0	178.173.199.255	Shiraz Hamyar Co.
178.173.200.0	178.173.200.255	Shiraz Hamyar Co.
178.173.201.0	178.173.201.255	Shiraz Hamyar Co.
178.173.202.0	178.173.202.255	Shiraz Hamyar Co.
178.173.203.0	178.173.203.255	Shiraz Hamyar Co.
178.173.204.0	178.173.204.255	Shiraz Hamyar Co.
178.173.205.0	178.173.205.255	Shiraz Hamyar Co.
178.173.206.0	178.173.206.255	Shiraz Hamyar Co.
178.173.207.0	178.173.207.255	Shiraz Hamyar Co.
178.173.208.0	178.173.208.255	Shiraz Hamyar Co.
178.173.209.0	178.173.209.255	Shiraz Hamyar Co.
178.173.210.0	178.173.210.255	Shiraz Hamyar Co.
178.173.211.0	178.173.211.255	Shiraz Hamyar Co.
178.173.212.0	178.173.212.255	Shiraz Hamyar Co.
178.173.213.0	178.173.213.255	Shiraz Hamyar Co.
178.173.214.0	178.173.214.255	Shiraz Hamyar Co.
178.173.215.0	178.173.215.255	Shiraz Hamyar Co.
178.173.216.0	178.173.216.255	Shiraz Hamyar Co.
178.173.217.0	178.173.217.255	Shiraz Hamyar Co.
178.173.218.0	178.173.218.255	Shiraz Hamyar Co.
178.173.219.0	178.173.219.255	Shiraz Hamyar Co.
178.173.220.0	178.173.220.255	Shiraz Hamyar Co.
178.173.221.0	178.173.221.255	Shiraz Hamyar Co.
178.173.222.0	178.173.222.255	Shiraz Hamyar Co.
178.173.223.0	178.173.223.255	Shiraz Hamyar Co.
178.211.145.0	178.211.145.255	Vandad Vira Hooman LLC
178.215.0.0	178.215.1.255	Islamic Republic of Iran Post PJSC
178.215.2.0	178.215.2.255	Islamic Republic of Iran Post PJSC
178.215.3.0	178.215.3.255	Islamic Republic of Iran Post PJSC
178.215.4.0	178.215.7.255	Islamic Republic of Iran Post PJSC
178.215.8.0	178.215.8.255	Islamic Republic of Iran Post PJSC
178.215.9.0	178.215.9.255	Islamic Republic of Iran Post PJSC
178.215.10.0	178.215.11.255	Islamic Republic of Iran Post PJSC
178.215.12.0	178.215.15.255	Islamic Republic of Iran Post PJSC
178.215.16.0	178.215.31.255	Islamic Republic of Iran Post PJSC
178.215.32.0	178.215.63.255	Islamic Republic of Iran Post PJSC
178.216.248.0	178.216.248.255	Asiatech Data Transmission Company
178.216.249.0	178.216.249.255	Asiatech Data Transmission Company
178.216.250.0	178.216.250.255	Asiatech Data Transmission Company
178.216.251.0	178.216.251.255	Asiatech Data Transmission Company
178.216.252.0	178.216.255.255	Iran Telecommunication Company Pjs
178.219.224.0	178.219.231.255	Islamic Republic of Iran Post PJSC
178.219.232.0	178.219.235.255	Islamic Republic of Iran Post PJSC
178.219.236.0	178.219.237.255	Islamic Republic of Iran Post PJSC
178.219.238.0	178.219.238.255	Islamic Republic of Iran Post PJSC
178.219.239.0	178.219.239.255	Islamic Republic of Iran Post PJSC
178.236.32.0	178.236.32.255	Toos-Ashena PJSC
178.236.33.0	178.236.33.255	Toos-Ashena PJSC
178.236.34.0	178.236.34.255	Toos-Ashena PJSC
178.236.35.0	178.236.35.255	Toos-Ashena PJSC
178.236.96.0	178.236.111.255	Iran Telecommunication Company Pjs
178.238.192.0	178.238.193.255	Iran Telecommunication Company Pjs
178.238.194.0	178.238.195.255	Iran Telecommunication Company Pjs
178.238.196.0	178.238.197.255	Iran Telecommunication Company Pjs
178.238.198.0	178.238.199.255	Iran Telecommunication Company Pjs
178.238.200.0	178.238.201.255	Iran Telecommunication Company Pjs
178.238.202.0	178.238.203.255	Iran Telecommunication Company Pjs
178.238.204.0	178.238.205.255	Iran Telecommunication Company Pjs
178.238.206.0	178.238.207.255	Iran Telecommunication Company Pjs
178.239.144.0	178.239.144.255	Peyman Ertebatat Pouya Company (Ltd)
178.239.145.0	178.239.145.255	Pars Abr Toseeh Ertebatat Ltd
178.239.146.0	178.239.146.255	Mizban Dadeh Pardazi Pasargad Ltd
178.239.147.0	178.239.147.255	Zarin Amol Gozar Technology Development Co. Ltd
178.239.148.0	178.239.148.255	Atrin Information & Communications Technology Company PJS
178.239.149.0	178.239.149.255	Peyman Ertebatat Pouya Company (Ltd)
178.239.150.0	178.239.150.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
178.239.151.0	178.239.151.255	Vandad Vira Hooman LLC
178.239.152.0	178.239.152.255	Azarakhsh Ava-E Ahvaz Co
178.239.153.0	178.239.153.255	Atrin Information & Communications Technology Company PJS
178.239.154.0	178.239.154.255	Shabake Tarh Madar Iranian Private Joint Stock Company
178.239.155.0	178.239.155.255	Nahor Hadish Design and Architecture Co. Ltd
178.239.156.0	178.239.156.255	Varesh Cloud Hosting Limited Liability Company
178.239.157.0	178.239.157.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
178.239.158.0	178.239.158.255	Vandad Vira Hooman LLC
178.239.159.0	178.239.159.255	Varesh Cloud Hosting Limited Liability Company
178.248.40.0	178.248.40.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
178.248.41.0	178.248.41.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
178.248.42.0	178.248.42.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
178.248.43.0	178.248.43.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
178.248.44.0	178.248.47.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
178.251.208.0	178.251.211.255	Telecommunication Infrastructure Company
178.251.212.0	178.251.212.255	Telecommunication Infrastructure Company
178.251.213.0	178.251.213.255	Telecommunication Infrastructure Company
178.251.214.0	178.251.214.255	Telecommunication Infrastructure Company
178.251.215.0	178.251.215.255	Telecommunication Infrastructure Company
178.252.128.0	178.252.135.255	GOSTARESH-E-ERTEBATAT-E MABNA COMPANY (Private Joint Stock)
178.252.136.0	178.252.137.255	GOSTARESH-E-ERTEBATAT-E MABNA COMPANY (Private Joint Stock)
178.252.138.0	178.252.138.255	GOSTARESH-E-ERTEBATAT-E MABNA COMPANY (Private Joint Stock)
178.252.139.0	178.252.139.255	GOSTARESH-E-ERTEBATAT-E MABNA COMPANY (Private Joint Stock)
178.252.140.0	178.252.143.255	GOSTARESH-E-ERTEBATAT-E MABNA COMPANY (Private Joint Stock)
178.252.144.0	178.252.147.255	GOSTARESH-E-ERTEBATAT-E MABNA COMPANY (Private Joint Stock)
178.252.148.0	178.252.149.255	GOSTARESH-E-ERTEBATAT-E MABNA COMPANY (Private Joint Stock)
178.252.150.0	178.252.150.255	AFTAB Communications and Informatics Private Joint Stock Company
178.252.151.0	178.252.151.255	GOSTARESH-E-ERTEBATAT-E MABNA COMPANY (Private Joint Stock)
178.252.152.0	178.252.159.255	GOSTARESH-E-ERTEBATAT-E MABNA COMPANY (Private Joint Stock)
178.252.160.0	178.252.191.255	GOSTARESH-E-ERTEBATAT-E MABNA COMPANY (Private Joint Stock)
185.1.77.0	185.1.77.255	Telecommunication Infrastructure Company
185.2.12.0	185.2.12.255	Faraso Samaneh Pasargad Ltd.
185.2.13.0	185.2.13.255	Faraso Samaneh Pasargad Ltd.
185.2.14.0	185.2.14.255	Faraso Samaneh Pasargad Ltd.
185.2.15.0	185.2.15.255	Faraso Samaneh Pasargad Ltd.
185.3.124.0	185.3.124.255	Dade Pardazi Mobinhost Co Ltd
185.3.125.0	185.3.125.255	Afranet
185.3.126.0	185.3.126.255	Afranet
185.3.127.0	185.3.127.255	Afranet
185.3.200.0	185.3.200.255	Padidar Technology PJSC
185.3.201.0	185.3.201.255	Padidar Technology PJSC
185.3.202.0	185.3.202.255	Padidar Technology PJSC
185.3.203.0	185.3.203.255	Padidar Technology PJSC
185.3.212.0	185.3.215.255	Pouya shabakeh Asr Co. (LTD.)
185.4.0.0	185.4.0.255	Pishtaze e Jadid Information Company (Ltd.)
185.4.1.0	185.4.1.255	Pishtaze e Jadid Information Company (Ltd.)
185.4.2.0	185.4.2.255	Pishtaze e Jadid Information Company (Ltd.)
185.4.3.0	185.4.3.255	Pishtaze e Jadid Information Company (Ltd.)
185.4.16.0	185.4.16.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
185.4.17.0	185.4.17.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
185.4.18.0	185.4.19.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
185.4.28.0	185.4.28.255	Green Web Samaneh Novin PJSC
185.4.29.0	185.4.29.255	Green Web Samaneh Novin PJSC
185.4.30.0	185.4.30.255	Green Web Samaneh Novin PJSC
185.4.31.0	185.4.31.255	Green Web Samaneh Novin PJSC
185.4.104.0	185.4.104.255	Rayan Hamafza Company Public Joint Stock
185.4.105.0	185.4.105.255	Rayan Hamafza Company Public Joint Stock
185.4.106.0	185.4.106.255	Rayan Hamafza Company Public Joint Stock
185.4.107.0	185.4.107.255	Rayan Hamafza Company Public Joint Stock
185.5.156.0	185.5.156.255	Mobile Communication Company of Iran Plc
185.5.157.0	185.5.157.255	Mobile Communication Company of Iran Plc
185.5.158.0	185.5.158.255	Mobile Communication Company of Iran Plc
185.5.159.0	185.5.159.255	Mobile Communication Company of Iran Plc
185.5.213.0	185.5.213.255	Pooyesh Axe Information and Communication Technology PJS
185.7.172.0	185.7.172.255	Mizban Dadeh Roham Co
185.7.212.0	185.7.212.255	Vandad Vira Hooman LLC
185.8.172.0	185.8.172.255	Vandad Vira Hooman LLC
185.8.173.0	185.8.173.255	Vandad Vira Hooman LLC
185.8.174.0	185.8.174.255	Vandad Vira Hooman LLC
185.8.175.0	185.8.175.255	Vandad Vira Hooman LLC
185.10.71.0	185.10.71.255	Parsan Lin Co. PJS
185.10.72.0	185.10.72.255	Roshangar Rayaneh Tehran Co. Ltd.
185.10.73.0	185.10.73.255	Roshangar Rayaneh Tehran Co. Ltd.
185.10.74.0	185.10.74.255	Roshangar Rayaneh Tehran Co. Ltd.
185.10.75.0	185.10.75.255	Roshangar Rayaneh Tehran Co. Ltd.
185.11.68.0	185.11.68.255	Pejvak Ertebatat Atiyeh Roshan Company (P.J.S.)
185.11.69.0	185.11.69.255	Pejvak Ertebatat Atiyeh Roshan Company (P.J.S.)
185.11.70.0	185.11.70.255	Pejvak Ertebatat Atiyeh Roshan Company (P.J.S.)
185.11.71.0	185.11.71.255	Pejvak Ertebatat Atiyeh Roshan Company (P.J.S.)
185.11.88.0	185.11.88.255	Telecommunication Infrastructure Company
185.11.89.0	185.11.89.255	Telecommunication Infrastructure Company
185.11.90.0	185.11.90.255	Telecommunication Infrastructure Company
185.11.91.0	185.11.91.255	Telecommunication Infrastructure Company
185.11.176.0	185.11.177.255	Khalij Fars Ettela Resan LTD
185.11.178.0	185.11.178.255	Khalij Fars Ettela Resan LTD
185.11.179.0	185.11.179.255	Khalij Fars Ettela Resan LTD
185.12.60.0	185.12.61.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
185.12.62.0	185.12.62.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
185.12.63.0	185.12.63.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
185.12.100.0	185.12.100.255	Naji Research & Development Company Private Join-Stock
185.12.101.0	185.12.101.255	Naji Research & Development Company Private Join-Stock
185.12.102.0	185.12.102.255	Naji Research & Development Company Private Join-Stock
185.12.103.0	185.12.103.255	Naji Research & Development Company Private Join-Stock
185.13.228.0	185.13.228.255	Parsan Lin Co. PJS
185.13.229.0	185.13.229.255	Parsan Lin Co. PJS
185.13.230.0	185.13.230.255	Parsan Lin Co. PJS
185.13.231.0	185.13.231.255	Parsan Lin Co. PJS
185.14.80.0	185.14.80.255	Iran Meteorological Organization
185.14.81.0	185.14.81.255	Iran Meteorological Organization
185.14.82.0	185.14.82.255	Iran Meteorological Organization
185.14.83.0	185.14.83.255	Iran Meteorological Organization
185.14.160.0	185.14.161.255	Respina Networks & Beyond PJSC
185.14.162.0	185.14.162.255	Respina Networks & Beyond PJSC
185.14.163.0	185.14.163.255	Respina Networks & Beyond PJSC
185.16.232.0	185.16.233.255	Petiak System Co JSC
185.16.234.0	185.16.235.255	Petiak System Co JSC
185.18.156.0	185.18.159.255	Iran Telecommunication Company Pjs
185.18.212.0	185.18.212.255	Sefroyek Pardaz Engineering PJSC
185.18.213.0	185.18.213.255	Sefroyek Pardaz Engineering PJSC
185.18.214.0	185.18.214.255	Sefroyek Pardaz Engineering PJSC
185.18.215.0	185.18.215.255	Sefroyek Pardaz Engineering PJSC
185.19.201.0	185.19.201.255	Vandad Vira Hooman LLC
185.20.160.0	185.20.160.255	Fanava Group
185.20.161.0	185.20.161.255	Fanava Group
185.20.162.0	185.20.162.255	Fanava Group
185.20.163.0	185.20.163.255	Fanava Group
185.21.68.0	185.21.71.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
185.21.76.0	185.21.79.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
185.22.28.0	185.22.28.255	Mobile Communication Company of Iran Plc
185.22.29.0	185.22.29.255	Mobile Communication Company of Iran Plc
185.22.30.0	185.22.31.255	Mobin Net Communication Company (Private Joint Stock)
185.23.128.0	185.23.128.255	Pars Data Processing Company (LTD)
185.23.129.0	185.23.129.255	Pars Data Processing Company (LTD)
185.23.130.0	185.23.130.255	Pars Data Processing Company (LTD)
185.23.131.0	185.23.131.255	Pars Data Processing Company (LTD)
185.24.136.0	185.24.136.255	Rightel Communication Service Company PJS
185.24.137.0	185.24.137.255	Rightel Communication Service Company PJS
185.24.138.0	185.24.138.255	Rightel Communication Service Company PJS
185.24.139.0	185.24.139.255	Rightel Communication Service Company PJS
185.24.151.0	185.24.151.255	Farhang Azma Communications Company LTD
185.24.228.0	185.24.231.255	Rightel Communication Service Company PJS
185.24.252.0	185.24.252.255	E-Money Net Developers 24 Company Private Joint Stock
185.24.253.0	185.24.253.255	E-Money Net Developers 24 Company Private Joint Stock
185.24.254.0	185.24.254.255	E-Money Net Developers 24 Company Private Joint Stock
185.24.255.0	185.24.255.255	E-Money Net Developers 24 Company Private Joint Stock
185.25.172.0	185.25.175.255	Damavand Electronic Card Company (Private J.S)
185.26.32.0	185.26.32.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
185.26.33.0	185.26.33.255	Patron Technology Persia Ltd
185.26.34.0	185.26.34.255	Varesh Cloud Hosting Limited Liability Company
185.26.35.0	185.26.35.255	Varesh Cloud Hosting Limited Liability Company
185.26.232.0	185.26.233.255	PJSC Badr Rayan Jonoob
185.26.234.0	185.26.235.255	PJSC Badr Rayan Jonoob
185.27.44.0	185.27.44.255	Web Dadeh Paydar Co (Ltd)
185.27.45.0	185.27.45.255	Giti Secure Cloud LLC
185.29.220.0	185.29.221.255	Farhang Azma Communications Company LTD
185.29.222.0	185.29.222.255	Farhang Azma Communications Company LTD
185.29.223.0	185.29.223.255	Peyman Ertebatat Pouya Company (Ltd)
185.30.4.0	185.30.7.255	Bank Ghavamin (Public Joint Stock)
185.30.76.0	185.30.79.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
185.31.8.0	185.31.8.255	Web Dadeh Paydar Co (Ltd)
185.31.124.0	185.31.124.255	Tourism Bank (Public Joint Stock)
185.31.125.0	185.31.125.255	Tourism Bank (Public Joint Stock)
185.31.126.0	185.31.126.255	Tourism Bank (Public Joint Stock)
185.31.127.0	185.31.127.255	Tourism Bank (Public Joint Stock)
185.32.128.0	185.32.131.255	Iranian Company of Saipa Automobile Manufacturing Public Joint Stock
185.33.25.0	185.33.25.255	Azarakhsh Ava-E Ahvaz Co
185.36.228.0	185.36.228.255	Sefroyek Pardaz Engineering PJSC
185.36.231.0	185.36.231.255	Noavaran Shabakeh Sabz Mehregan (Ltd.)
185.37.52.0	185.37.52.255	Tebyan-e-Noor Cultural-Artistic Institute
185.37.53.0	185.37.53.255	Tebyan-e-Noor Cultural-Artistic Institute
185.37.54.0	185.37.54.255	Tebyan-e-Noor Cultural-Artistic Institute
185.37.55.0	185.37.55.255	Tebyan-e-Noor Cultural-Artistic Institute
185.40.16.0	185.40.16.255	Noavaran Shabakeh Sabz Mehregan (Ltd.)
185.40.240.0	185.40.243.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
185.41.0.0	185.41.0.255	Dade Pardaz Kimia Pouyesh PJS.
185.41.1.0	185.41.1.255	Dade Pardaz Kimia Pouyesh PJS.
185.41.2.0	185.41.2.255	Dade Pardaz Kimia Pouyesh PJS.
185.41.3.0	185.41.3.255	Dade Pardaz Kimia Pouyesh PJS.
185.41.220.0	185.41.220.255	Central Securities Depository and Funds Settlement Public JSC
185.41.221.0	185.41.221.255	Central Securities Depository and Funds Settlement Public JSC
185.41.222.0	185.41.222.255	Central Securities Depository and Funds Settlement Public JSC
185.41.223.0	185.41.223.255	Central Securities Depository and Funds Settlement Public JSC
185.42.212.0	185.42.212.255	Atrin Information & Communications Technology Company PJS
185.42.213.0	185.42.213.255	Rayaneh Gostar Farzanegan Ahwaz Company LTD.
185.42.214.0	185.42.214.255	Rayaneh Gostar Farzanegan Ahwaz Company LTD.
185.42.215.0	185.42.215.255	Rayaneh Gostar Farzanegan Ahwaz Company LTD.
185.42.224.0	185.42.224.255	Pardazeshgar Ray Azma Co. Ltd.
185.42.225.0	185.42.225.255	Pardazeshgar Ray Azma Co. Ltd.
185.42.226.0	185.42.226.255	Pardazeshgar Ray Azma Co. Ltd.
185.42.227.0	185.42.227.255	Pardazeshgar Ray Azma Co. Ltd.
185.44.36.0	185.44.36.255	Manzoomeh Negaran Co. (PJS)
185.44.37.0	185.44.37.255	Manzoomeh Negaran Co. (PJS)
185.44.38.0	185.44.39.255	Manzoomeh Negaran Co. (PJS)
185.44.100.0	185.44.100.255	Melal Credit Company (Public Joint Stock)
185.44.101.0	185.44.101.255	Melal Credit Company (Public Joint Stock)
185.44.102.0	185.44.102.255	Melal Credit Company (Public Joint Stock)
185.44.103.0	185.44.103.255	Melal Credit Company (Public Joint Stock)
185.44.112.0	185.44.112.255	Gostaresh Etrtebatate Apadana Ltd.
185.44.113.0	185.44.113.255	Gostaresh Etrtebatate Apadana Ltd.
185.44.114.0	185.44.114.255	Gostaresh Etrtebatate Apadana Ltd.
185.44.115.0	185.44.115.255	Gostaresh Etrtebatate Apadana Ltd.
185.45.188.0	185.45.191.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
185.46.0.0	185.46.3.255	Parvaresh Dadeha Co. Private Joint Stock
185.46.108.0	185.46.111.255	Pooya Parto Qeshm Cooperative Company
185.46.216.0	185.46.216.255	Shabakeh Gostar Shahriyar Co. (Ltd.)
185.46.217.0	185.46.217.255	Shabakeh Gostar Shahriyar Co. (Ltd.)
185.46.218.0	185.46.218.255	Shabakeh Gostar Shahriyar Co. (Ltd.)
185.46.219.0	185.46.219.255	Shabakeh Gostar Shahriyar Co. (Ltd.)
185.47.48.0	185.47.48.255	Hesabgar Pardaz Gharb PJSC
185.47.49.0	185.47.49.255	Hesabgar Pardaz Gharb PJSC
185.47.50.0	185.47.50.255	Hesabgar Pardaz Gharb PJSC
185.47.51.0	185.47.51.255	Hesabgar Pardaz Gharb PJSC
185.49.84.0	185.49.84.255	Asiatech Data Transmission Company
185.49.85.0	185.49.85.255	Asiatech Data Transmission Company
185.49.86.0	185.49.86.255	Asiatech Data Transmission Company
185.49.87.0	185.49.87.255	Asiatech Data Transmission Company
185.49.96.0	185.49.96.255	Rayaneh Pardazan Baran Co. Ltd.
185.49.97.0	185.49.97.255	Rayaneh Pardazan Baran Co. Ltd.
185.49.98.0	185.49.98.255	Rayaneh Pardazan Baran Co. Ltd.
185.49.99.0	185.49.99.255	Rayaneh Pardazan Baran Co. Ltd.
185.49.104.0	185.49.105.255	Toos-Ashena PJSC
185.49.106.0	185.49.107.255	Mobin Net Communication Company (Private Joint Stock)
185.49.174.0	185.49.174.255	Darvag Cloud Infrastructure Innovators Co.pjs
185.49.231.0	185.49.231.255	Saba Abr Mizban LLC
185.50.36.0	185.50.36.255	Rayanmehr Danesh Sanj Company Ltd
185.50.37.0	185.50.37.255	Green Web Samaneh Novin PJSC
185.50.38.0	185.50.38.255	Green Web Samaneh Novin PJSC
185.50.39.0	185.50.39.255	Green Web Samaneh Novin PJSC
185.51.40.0	185.51.40.255	Araax Dadeh Gostar Information and Communication Development Co (Private Joint Stock)
185.51.41.0	185.51.41.255	Araax Dadeh Gostar Information and Communication Development Co (Private Joint Stock)
185.51.42.0	185.51.42.255	Araax Dadeh Gostar Information and Communication Development Co (Private Joint Stock)
185.51.43.0	185.51.43.255	Araax Dadeh Gostar Information and Communication Development Co (Private Joint Stock)
185.51.200.0	185.51.200.255	Sefroyek Pardaz Engineering PJSC
185.51.201.0	185.51.201.255	Sefroyek Pardaz Engineering PJSC
185.51.202.0	185.51.202.255	Sefroyek Pardaz Engineering PJSC
185.51.203.0	185.51.203.255	Sefroyek Pardaz Engineering PJSC
185.53.140.0	185.53.140.255	Dade Pardazi Mobinhost Co Ltd
185.53.141.0	185.53.141.255	Dade Pardazi Mobinhost Co Ltd
185.53.142.0	185.53.142.255	Dade Pardazi Mobinhost Co Ltd
185.53.143.0	185.53.143.255	Dade Pardazi Mobinhost Co Ltd
185.55.224.0	185.55.224.255	Fanavari Serverpars Argham Gostar Company Ltd.
185.55.225.0	185.55.225.255	Fanavari Serverpars Argham Gostar Company Ltd.
185.55.226.0	185.55.226.255	Fanavari Serverpars Argham Gostar Company Ltd.
185.55.227.0	185.55.227.255	Fanavari Serverpars Argham Gostar Company Ltd.
185.56.92.0	185.56.92.255	Naroon Intelligent Communications Ltd.
185.56.93.0	185.56.93.255	Naroon Intelligent Communications Ltd.
185.56.94.0	185.56.94.255	Naroon Intelligent Communications Ltd.
185.56.95.0	185.56.95.255	Naroon Intelligent Communications Ltd.
185.56.96.0	185.56.97.255	Andishe Sabz Khazar Co. P.J.S.
185.56.98.0	185.56.98.255	Andishe Sabz Khazar Co. P.J.S.
185.56.99.0	185.56.99.255	Andishe Sabz Khazar Co. P.J.S.
185.57.132.0	185.57.133.255	Bozorg Net-E Aria
185.57.134.0	185.57.135.255	Bozorg Net-E Aria
185.57.164.0	185.57.164.255	Asr-e Enteghal-e Dadeha Company (Private J.S.)
185.57.165.0	185.57.165.255	Asr-e Enteghal-e Dadeha Company (Private J.S.)
185.57.166.0	185.57.166.255	Asr-e Enteghal-e Dadeha Company (Private J.S.)
185.57.167.0	185.57.167.255	Asr-e Enteghal-e Dadeha Company (Private J.S.)
185.57.200.0	185.57.203.255	Iran Information Technology Company PJSC
185.58.240.0	185.58.240.255	Sefroyek Pardaz Engineering PJSC
185.58.241.0	185.58.241.255	Sefroyek Pardaz Engineering PJSC
185.58.242.0	185.58.242.255	Sefroyek Pardaz Engineering PJSC
185.58.243.0	185.58.243.255	Sefroyek Pardaz Engineering PJSC
185.59.112.0	185.59.112.255	Pejvak Ertebatat Atiyeh Roshan Company (P.J.S.)
185.59.113.0	185.59.113.255	Pejvak Ertebatat Atiyeh Roshan Company (P.J.S.)
185.60.32.0	185.60.32.255	Pardakht Electronic Pasargad (P.J.S)
185.60.33.0	185.60.33.255	Pardakht Electronic Pasargad (P.J.S)
185.60.34.0	185.60.34.255	Pardakht Electronic Pasargad (P.J.S)
185.60.35.0	185.60.35.255	Pardakht Electronic Pasargad (P.J.S)
185.60.59.0	185.60.59.255	Taymaz Gostaran Naghshe Almas Co. ( Ltd )
185.60.136.0	185.60.136.255	Soroush Rasanheh Company Ltd
185.60.137.0	185.60.137.255	Soroush Rasanheh Company Ltd
185.60.138.0	185.60.138.255	Soroush Rasanheh Company Ltd
185.60.139.0	185.60.139.255	Soroush Rasanheh Company Ltd
185.62.232.0	185.62.235.255	Institute for Research in Fundamental Sciences
185.63.113.0	185.63.113.255	Avaye Segal Net Co. PJS
185.63.114.0	185.63.114.255	Sindad Network Technology PJSC
185.63.236.0	185.63.236.255	Iranian Information Technology PJSC
185.63.237.0	185.63.237.255	Iranian Information Technology PJSC
185.63.238.0	185.63.239.255	Iranian Information Technology Company PLC
185.64.176.0	185.64.177.255	Shiraz University
185.64.178.0	185.64.179.255	Shiraz University
185.65.118.0	185.65.118.255	Sabz Pardazesh System Ghadir Co. (P.J.S.)
185.66.224.0	185.66.224.255	Pooya Parto Qeshm Cooperative Company
185.66.225.0	185.66.225.255	Pooya Parto Qeshm Cooperative Company
185.66.226.0	185.66.226.255	Pooya Parto Qeshm Cooperative Company
185.66.227.0	185.66.227.255	Respina Networks & Beyond PJSC
185.66.228.0	185.66.228.255	Lamerd Information & Communication Technology Co. ltd
185.66.229.0	185.66.229.255	Lamerd Information & Communication Technology Co. ltd
185.66.230.0	185.66.230.255	Web Gostaran Bandar Company (PJS)
185.66.231.0	185.66.231.255	Lamerd Information & Communication Technology Co. ltd
185.67.12.0	185.67.12.255	University of Zanjan
185.67.13.0	185.67.13.255	University of Zanjan
185.67.14.0	185.67.14.255	University of Zanjan
185.67.15.0	185.67.15.255	University of Zanjan
185.67.100.0	185.67.101.255	Asiatech Data Transmission Company
185.67.102.0	185.67.103.255	Asiatech Data Transmission Company
185.67.156.0	185.67.156.255	Islamic Azad University
185.67.157.0	185.67.157.255	Islamic Azad University
185.67.158.0	185.67.158.255	Islamic Azad University
185.67.159.0	185.67.159.255	Islamic Azad University
185.67.212.0	185.67.212.255	Persia Telecom Company
185.67.213.0	185.67.213.255	Persia Telecom Company
185.67.214.0	185.67.214.255	Persia Telecom Company
185.67.215.0	185.67.215.255	Persia Telecom Company
185.69.108.0	185.69.109.255	Tehran Municipality ICT Organization
185.69.110.0	185.69.111.255	Tehran Municipality ICT Organization
185.70.60.0	185.70.63.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
185.71.152.0	185.71.155.255	Iran Telecommunication Company Pjs
185.71.192.0	185.71.193.255	Pishgaman Toseeh Fanavari Etelaat Va Ertebatat Jonoub (Joint Stock Company)
185.71.194.0	185.71.195.255	Pishgaman Toseeh Fanavari Etelaat Va Ertebatat Jonoub (Joint Stock Company)
185.72.24.0	185.72.25.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
185.72.26.0	185.72.26.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
185.72.27.0	185.72.27.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
185.72.80.0	185.72.83.255	Iran Telecommunication Company Pjs
185.73.0.0	185.73.0.255	Aria Shatel PJSC
185.73.1.0	185.73.1.255	Aria Shatel PJSC
185.73.2.0	185.73.3.255	Aria Shatel PJSC
185.73.76.0	185.73.76.255	Bank of Industry and Mine PJSC
185.73.77.0	185.73.77.255	Bank of Industry and Mine PJSC
185.73.78.0	185.73.78.255	Bank of Industry and Mine PJSC
185.73.79.0	185.73.79.255	Bank of Industry and Mine PJSC
185.73.112.0	185.73.112.255	Negah Roshan Pars Company (PJS)
185.73.114.0	185.73.114.255	Negah Roshan Pars Company (PJS)
185.73.226.0	185.73.226.255	Netafraz Iranian Ltd.
185.74.164.0	185.74.164.255	IR Iran Civil Aviation Authority
185.74.165.0	185.74.165.255	Noyan Abr Arvan Co. ( Private Joint Stock)
185.74.166.0	185.74.166.255	IR Iran Civil Aviation Authority
185.74.167.0	185.74.167.255	IR Iran Civil Aviation Authority
185.74.221.0	185.74.221.255	Sindad Network Technology PJSC
185.75.196.0	185.75.196.255	Sepehr Saye Ban Shokoufaei Company (LTD)
185.75.197.0	185.75.197.255	Sepehr Saye Ban Shokoufaei Company (LTD)
185.75.198.0	185.75.198.255	Sepehr Saye Ban Shokoufaei Company (LTD)
185.75.199.0	185.75.199.255	Sepehr Saye Ban Shokoufaei Company (LTD)
185.75.204.0	185.75.204.255	Cooperative Afra ertebatat-e-sabet-e Rasa Co
185.75.205.0	185.75.205.255	Cooperative Afra ertebatat-e-sabet-e Rasa Co
185.75.206.0	185.75.206.255	Cooperative Afra ertebatat-e-sabet-e Rasa Co
185.75.207.0	185.75.207.255	Cooperative Afra ertebatat-e-sabet-e Rasa Co
185.76.248.0	185.76.249.255	Tehran Municipality ICT Organization
185.76.250.0	185.76.251.255	Tehran Municipality ICT Organization
185.77.21.0	185.77.21.255	Pouyasazan Fanavari-Ye Ettela'at LLC
185.78.20.0	185.78.20.255	Tebyan-e-Noor Cultural-Artistic Institute
185.78.21.0	185.78.21.255	Tebyan-e-Noor Cultural-Artistic Institute
185.78.22.0	185.78.22.255	Tebyan-e-Noor Cultural-Artistic Institute
185.78.23.0	185.78.23.255	Tebyan-e-Noor Cultural-Artistic Institute
185.79.16.0	185.79.16.255	Charisma Financial Information Processing PJSC
185.79.60.0	185.79.60.255	Kish Iran Credit Card Corporation (Public Corporation)
185.79.61.0	185.79.61.255	Kish Iran Credit Card Corporation (Public Corporation)
185.79.62.0	185.79.63.255	Kish Iran Credit Card Corporation (Public Corporation)
185.79.96.0	185.79.96.255	Asiatech Data Transmission Company
185.79.97.0	185.79.97.255	Asiatech Data Transmission Company
185.79.98.0	185.79.98.255	Asiatech Data Transmission Company
185.79.99.0	185.79.99.255	Asiatech Data Transmission Company
185.79.156.0	185.79.159.255	Iran Telecommunication Company Pjs
185.80.100.0	185.80.103.255	Iran Telecommunication Company Pjs
185.80.198.0	185.80.198.255	Fannavaran-e Idea Pardaz-e Saba PJSC
185.80.199.0	185.80.199.255	Fannavaran-e Idea Pardaz-e Saba PJSC
185.81.40.0	185.81.43.255	Rooyekhat Media Company Ltd
185.81.96.0	185.81.96.255	Mizban Dade Pasargad LLC
185.81.97.0	185.81.97.255	Mizban Dade Pasargad LLC
185.81.98.0	185.81.98.255	Mizban Dade Pasargad LLC
185.81.99.0	185.81.99.255	Mizban Dade Pasargad LLC
185.82.28.0	185.82.28.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
185.82.29.0	185.82.29.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
185.82.30.0	185.82.31.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
185.82.64.0	185.82.64.255	Hesabgar Pardaz Gharb PJSC
185.82.65.0	185.82.65.255	Hesabgar Pardaz Gharb PJSC
185.82.66.0	185.82.66.255	Hesabgar Pardaz Gharb PJSC
185.82.67.0	185.82.67.255	Hesabgar Pardaz Gharb PJSC
185.82.136.0	185.82.139.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
185.82.164.0	185.82.165.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
185.82.166.0	185.82.167.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
185.82.180.0	185.82.183.255	Maskan Bank
185.83.28.0	185.83.28.255	Rayaneh Gostar Farzanegan Ahwaz Company LTD.
185.83.29.0	185.83.29.255	Rayaneh Gostar Farzanegan Ahwaz Company LTD.
185.83.30.0	185.83.30.255	Rayaneh Gostar Farzanegan Ahwaz Company LTD.
185.83.31.0	185.83.31.255	Rayaneh Gostar Farzanegan Ahwaz Company LTD.
185.83.76.0	185.83.79.255	Iranian Research Organization for Science & Technology
185.83.80.0	185.83.81.255	Datis Arian Gheshm Software Company PJSC
185.83.82.0	185.83.82.255	Datis Arian Gheshm Software Company PJSC
185.83.83.0	185.83.83.255	Datis Arian Gheshm Software Company PJSC
185.83.88.0	185.83.88.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
185.83.89.0	185.83.89.255	Rayankadeh Apadana Company Ltd
185.83.90.0	185.83.91.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
185.83.112.0	185.83.112.255	Noavaran Shabakeh Sabz Mehregan (Ltd.)
185.83.113.0	185.83.113.255	Noavaran Shabakeh Sabz Mehregan (Ltd.)
185.83.114.0	185.83.114.255	Noavaran Shabakeh Sabz Mehregan (Ltd.)
185.83.115.0	185.83.115.255	Noavaran Shabakeh Sabz Mehregan (Ltd.)
185.83.180.0	185.83.180.255	Cooperative Afra ertebatat-e-sabet-e Rasa Co
185.83.181.0	185.83.181.255	Noavaran Shabakeh Sabz Mehregan (Ltd.)
185.83.182.0	185.83.182.255	Taymaz Gostaran Naghshe Almas Co. ( Ltd )
185.83.183.0	185.83.183.255	Noavaran Shabakeh Sabz Mehregan (Ltd.)
185.83.184.0	185.83.187.255	Rayaneh Gostar Farzanegan Ahwaz Company LTD.
185.83.196.0	185.83.196.255	Pejvak Ertebatat Atiyeh Roshan Company (P.J.S.)
185.83.197.0	185.83.197.255	Pejvak Ertebatat Atiyeh Roshan Company (P.J.S.)
185.83.198.0	185.83.198.255	Pejvak Ertebatat Atiyeh Roshan Company (P.J.S.)
185.83.199.0	185.83.199.255	Pejvak Ertebatat Atiyeh Roshan Company (P.J.S.)
185.83.208.0	185.83.208.255	Behin Ertebatat Faragir Co. Ltd
185.83.209.0	185.83.209.255	Behin Ertebatat Faragir Co. Ltd
185.83.210.0	185.83.210.255	Behin Ertebatat Faragir Co. Ltd
185.83.211.0	185.83.211.255	Behin Ertebatat Faragir Co. Ltd
185.84.158.0	185.84.158.255	Ettelaat Fanavarn-E Tookan Co Ltd
185.84.220.0	185.84.221.255	Ferdowsi University of Mashhad
185.84.222.0	185.84.223.255	Ferdowsi University of Mashhad
185.84.226.0	185.84.226.255	Zirsakht Pishro Abr Damavand Private Joint Stock Company
185.85.68.0	185.85.71.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
185.85.136.0	185.85.139.255	Ahang Rayan Malayer Co. LTD
185.86.36.0	185.86.36.255	Ravand Tazeh Co .Pjs.
185.86.37.0	185.86.37.255	Ravand Tazeh Co .Pjs.
185.86.38.0	185.86.38.255	Ravand Tazeh Co .Pjs.
185.86.39.0	185.86.39.255	Ravand Tazeh Co .Pjs.
185.86.180.0	185.86.180.255	Sindad Network Technology PJSC
185.86.181.0	185.86.181.255	Sindad Network Technology PJSC
185.86.182.0	185.86.182.255	Web Gostaran Bandar Company (PJS)
185.86.183.0	185.86.183.255	Web Gostaran Bandar Company (PJS)
185.88.48.0	185.88.49.255	Iran Telecommunication Company Pjs
185.88.50.0	185.88.51.255	Iran Telecommunication Company Pjs
185.88.152.0	185.88.152.255	Parsian High Tech Company PJSC
185.88.153.0	185.88.153.255	Parsian High Tech Company PJSC
185.88.154.0	185.88.154.255	Parsian High Tech Company PJSC
185.88.155.0	185.88.155.255	Parsian High Tech Company PJSC
185.88.176.0	185.88.176.255	Fannavari Etelaate Samane Kavoshgar Ide Ltd.
185.88.177.0	185.88.177.255	Fannavari Etelaate Samane Kavoshgar Ide Ltd.
185.88.178.0	185.88.178.255	Fannavari Etelaate Samane Kavoshgar Ide Ltd.
185.88.179.0	185.88.179.255	Fannavari Etelaate Samane Kavoshgar Ide Ltd.
185.88.252.0	185.88.253.255	Orum Torkan Rayaneh LLC
185.88.254.0	185.88.255.255	Orum Torkan Rayaneh LLC
185.89.22.0	185.89.22.255	Noavaran Shabakeh Sabz Mehregan (Ltd.)
185.89.112.0	185.89.112.255	Pardazesh Rayaneh Bita Co.Ltd
185.89.113.0	185.89.113.255	Pardazesh Rayaneh Bita Co.Ltd
185.89.114.0	185.89.114.255	Pardazesh Rayaneh Bita Co.Ltd
185.89.115.0	185.89.115.255	Pardazesh Rayaneh Bita Co.Ltd
185.92.4.0	185.92.7.255	Rasaneh Mehr Vatan Co. PJS
185.92.8.0	185.92.8.255	Iran Insurance Joint Stock Company
185.92.9.0	185.92.9.255	Iran Insurance Joint Stock Company
185.92.10.0	185.92.10.255	Iran Insurance Joint Stock Company
185.92.11.0	185.92.11.255	Iran Insurance Joint Stock Company
185.92.40.0	185.92.43.255	Rayan Ravesh Sena Ltd
185.93.88.0	185.93.88.255	Web Dadeh Paydar Co (Ltd)
185.94.96.0	185.94.96.255	Netmihan Communication Company Ltd
185.94.97.0	185.94.97.255	Netmihan Communication Company Ltd
185.94.98.0	185.94.98.255	Netmihan Communication Company Ltd
185.94.99.0	185.94.99.255	Netmihan Communication Company Ltd
185.94.180.0	185.94.180.255	Web Dadeh Paydar Co (Ltd)
185.94.181.0	185.94.181.255	Web Dadeh Paydar Co (Ltd)
185.95.60.0	185.95.63.255	Pishgaman Kavir Asia Private Joint Stock
185.95.152.0	185.95.155.255	Data Pardaz Afraz Private Joint Stock Company
185.95.180.0	185.95.180.255	Allameh Tabataba'i University
185.95.181.0	185.95.181.255	Allameh Tabataba'i University
185.95.182.0	185.95.182.255	Allameh Tabataba'i University
185.95.183.0	185.95.183.255	Allameh Tabataba'i University
185.96.240.0	185.96.240.255	Yazd University
185.96.241.0	185.96.241.255	Yazd University
185.96.242.0	185.96.242.255	Yazd University
185.96.243.0	185.96.243.255	Yazd University
185.97.116.0	185.97.119.255	Noyan Abr Arvan Co. ( Private Joint Stock)
185.98.112.0	185.98.112.255	Asiatech Data Transmission Company
185.98.113.0	185.98.113.255	Asiatech Data Transmission Company
185.98.114.0	185.98.114.255	Asiatech Data Transmission Company
185.98.115.0	185.98.115.255	Asiatech Data Transmission Company
185.99.212.0	185.99.213.255	Asiatech Data Transmission Company
185.99.214.0	185.99.215.255	Asiatech Data Transmission Company
185.100.44.0	185.100.47.255	Fara Ertebat Mabna PJS
185.101.39.0	185.101.39.255	Pishgam Rayan Dadeh Pajooh PJSC
185.101.228.0	185.101.229.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
185.101.230.0	185.101.230.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
185.101.231.0	185.101.231.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
185.103.84.0	185.103.84.255	Mobin Net Communication Company (Private Joint Stock)
185.103.85.0	185.103.85.255	Mobin Net Communication Company (Private Joint Stock)
185.103.86.0	185.103.87.255	Mobin Net Communication Company (Private Joint Stock)
185.103.128.0	185.103.131.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
185.103.201.0	185.103.201.255	Amn Electronic Kashef Management Co. PJSC
185.103.244.0	185.103.247.255	PJSC Badr Rayan Jonoob
185.103.248.0	185.103.251.255	PJSC Badr Rayan Jonoob
185.104.228.0	185.104.228.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
185.104.229.0	185.104.229.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
185.104.230.0	185.104.231.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
185.104.232.0	185.104.235.255	PJSC Badr Rayan Jonoob
185.104.240.0	185.104.243.255	PJSC Badr Rayan Jonoob
185.105.100.0	185.105.103.255	Iran Telecommunication Company Pjs
185.105.120.0	185.105.123.255	Shahrekord University
185.105.184.0	185.105.184.255	Afagh Andish Dadeh Pardis Co. Ltd
185.105.185.0	185.105.185.255	Afagh Andish Dadeh Pardis Co. Ltd
185.105.186.0	185.105.186.255	Afagh Andish Dadeh Pardis Co. Ltd
185.105.187.0	185.105.187.255	Afagh Andish Dadeh Pardis Co. Ltd
185.105.236.0	185.105.236.255	Patron Technology Persia Ltd
185.105.237.0	185.105.237.255	Amnpardaz Soft Corporation
185.105.238.0	185.105.238.255	Arman White Water Way PJS
185.105.239.0	185.105.239.255	Vandad Vira Hooman LLC
185.106.136.0	185.106.139.255	PJSC Badr Rayan Jonoob
185.106.144.0	185.106.147.255	PJSC Badr Rayan Jonoob
185.106.200.0	185.106.200.255	Netafraz Iranian Ltd.
185.106.201.0	185.106.201.255	Netafraz Iranian Ltd.
185.106.202.0	185.106.203.255	Netafraz Iranian Ltd.
185.106.228.0	185.106.229.255	Iran Telecommunication Company Pjs
185.106.230.0	185.106.231.255	Iran Telecommunication Company Pjs
185.107.28.0	185.107.28.255	PJSC Badr Rayan Jonoob
185.107.29.0	185.107.29.255	PJSC Badr Rayan Jonoob
185.107.30.0	185.107.31.255	PJSC Badr Rayan Jonoob
185.107.32.0	185.107.33.255	Iran Telecommunication Company Pjs
185.107.34.0	185.107.35.255	Iran Telecommunication Company Pjs
185.107.244.0	185.107.244.255	Araax Dadeh Gostar Information and Communication Development Co (Private Joint Stock)
185.107.245.0	185.107.245.255	Ava Mehr Informatics Company (Private Joint Stock)
185.107.246.0	185.107.246.255	Araax Dadeh Gostar Information and Communication Development Co (Private Joint Stock)
185.107.247.0	185.107.247.255	Araax Dadeh Gostar Information and Communication Development Co (Private Joint Stock)
185.107.248.0	185.107.251.255	Shahrekord University
185.108.96.0	185.108.99.255	Iran Telecommunication Company Pjs
185.108.164.0	185.108.167.255	PJSC Badr Rayan Jonoob
185.109.60.0	185.109.60.255	Pishgaman Toseeh Fanavari Etelaat Va Ertebatat Jonoub (Joint Stock Company)
185.109.61.0	185.109.61.255	Pishgaman Toseeh Fanavari Etelaat Va Ertebatat Jonoub (Joint Stock Company)
185.109.62.0	185.109.63.255	Pishgaman Toseeh Fanavari Etelaat Va Ertebatat Jonoub (Joint Stock Company)
185.109.72.0	185.109.72.255	Pishgaman Toseeh Fanavari Etelaat Va Ertebatat Jonoub (Joint Stock Company)
185.109.73.0	185.109.73.255	Pishgaman Toseeh Fanavari Etelaat Va Ertebatat Jonoub (Joint Stock Company)
185.109.74.0	185.109.74.255	Pishgaman Toseeh Fanavari Etelaat Va Ertebatat Jonoub (Joint Stock Company)
185.109.75.0	185.109.75.255	Pishgaman Toseeh Fanavari Etelaat Va Ertebatat Jonoub (Joint Stock Company)
185.109.80.0	185.109.81.255	Pishgaman Toseeh Fanavari Etelaat Va Ertebatat Jonoub (Joint Stock Company)
185.109.82.0	185.109.83.255	Pishgaman Toseeh Fanavari Etelaat Va Ertebatat Jonoub (Joint Stock Company)
185.109.128.0	185.109.131.255	IsIran
185.109.244.0	185.109.244.255	Asiatech Data Transmission Company
185.109.245.0	185.109.245.255	Asiatech Data Transmission Company
185.109.246.0	185.109.247.255	Asiatech Data Transmission Company
185.109.248.0	185.109.248.255	Asiatech Data Transmission Company
185.109.249.0	185.109.249.255	Asiatech Data Transmission Company
185.109.250.0	185.109.251.255	Asiatech Data Transmission Company
185.110.28.0	185.110.31.255	Respina Networks & Beyond PJSC
185.110.218.0	185.110.218.255	Bartar Andishan Avaye Bandar PJSC
185.110.219.0	185.110.219.255	Pooya Parto Qeshm Cooperative Company
185.110.228.0	185.110.229.255	Pars Car International Communication Age PJSC
185.110.230.0	185.110.230.255	Pars Car International Communication Age PJSC
185.110.231.0	185.110.231.255	Pars Car International Communication Age PJSC
185.110.236.0	185.110.236.255	Gostaresh Dade Ako Fanavaran Co LLC
185.110.237.0	185.110.237.255	Gostaresh Dade Ako Fanavaran Co LLC
185.110.238.0	185.110.238.255	Gostaresh Dade Ako Fanavaran Co LLC
185.110.239.0	185.110.239.255	Gostaresh Dade Ako Fanavaran Co LLC
185.110.244.0	185.110.244.255	Aria Shatel PJSC
185.110.245.0	185.110.245.255	Aria Shatel PJSC
185.110.246.0	185.110.246.255	Aria Shatel PJSC
185.110.247.0	185.110.247.255	Aria Shatel PJSC
185.110.252.0	185.110.252.255	SINA ATINET Company (Ltd)
185.110.253.0	185.110.253.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
185.110.254.0	185.110.254.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
185.110.255.0	185.110.255.255	SINA ATINET Company (Ltd)
185.111.8.0	185.111.11.255	Pishgaman Toseeh Fanavari Etelaat Va Ertebatat Jonoub (Joint Stock Company)
185.111.12.0	185.111.15.255	Pishgaman Toseeh Fanavari Etelaat Va Ertebatat Jonoub (Joint Stock Company)
185.111.64.0	185.111.65.255	Pishgaman Toseeh Fanavari Etelaat Va Ertebatat Jonoub (Joint Stock Company)
185.111.66.0	185.111.66.255	Pishgaman Toseeh Fanavari Etelaat Va Ertebatat Jonoub (Joint Stock Company)
185.111.67.0	185.111.67.255	Pishgaman Toseeh Fanavari Etelaat Va Ertebatat Jonoub (Joint Stock Company)
185.111.80.0	185.111.83.255	Iran Telecommunication Company Pjs
185.111.136.0	185.111.139.255	Pishgaman Toseeh Fanavari Etelaat Va Ertebatat Jonoub (Joint Stock Company)
185.112.32.0	185.112.32.255	Asiatech Data Transmission Company
185.112.33.0	185.112.33.255	Asiatech Data Transmission Company
185.112.34.0	185.112.34.255	Asiatech Data Transmission Company
185.112.35.0	185.112.35.255	Asiatech Data Transmission Company
185.112.36.0	185.112.39.255	Asiatech Data Transmission Company
185.112.128.0	185.112.131.255	Iran Telecommunication Company Pjs
185.112.148.0	185.112.148.255	Azarakhsh Ava-E Ahvaz Co
185.112.149.0	185.112.149.255	Sefroyek Pardaz Engineering PJSC
185.112.150.0	185.112.150.255	Sefroyek Pardaz Engineering PJSC
185.112.151.0	185.112.151.255	Sefroyek Pardaz Engineering PJSC
185.112.168.0	185.112.171.255	Pishgaman Toseeh Fanavari Etelaat Va Ertebatat Jonoub (Joint Stock Company)
185.113.9.0	185.113.9.255	Noavaran Shabakeh Sabz Mehregan (Ltd.)
185.113.56.0	185.113.59.255	Rayankadeh Apadana Company Ltd
185.113.112.0	185.113.112.255	University of Kashan
185.113.113.0	185.113.113.255	University of Kashan
185.113.114.0	185.113.114.255	University of Kashan
185.113.115.0	185.113.115.255	University of Kashan
185.113.248.0	185.113.248.255	Web Dadeh Paydar Co (Ltd)
185.114.188.0	185.114.191.255	Imam Khomeini International University of Qazvin
185.115.76.0	185.115.77.255	Iran Telecommunication Company Pjs
185.115.78.0	185.115.78.255	Iran Telecommunication Company Pjs
185.115.79.0	185.115.79.255	Iran Cell Service and Communication Company
185.115.148.0	185.115.149.255	GOSTARESH-E-ERTEBATAT-E MABNA COMPANY (Private Joint Stock)
185.115.150.0	185.115.150.255	Rahpooyan Pardazesh Gostar Sahra Co. PJS
185.115.151.0	185.115.151.255	Rahpooyan Pardazesh Gostar Sahra Co. PJS
185.115.168.0	185.115.168.255	Pars Data Processing Company (LTD)
185.115.169.0	185.115.169.255	Pars Data Processing Company (LTD)
185.115.170.0	185.115.170.255	Pars Data Processing Company (LTD)
185.115.171.0	185.115.171.255	Pars Data Processing Company (LTD)
185.116.20.0	185.116.20.255	Tarahan Shabake Sharif LTD
185.116.21.0	185.116.21.255	Tarahan Shabake Sharif LTD
185.116.22.0	185.116.22.255	Tarahan Shabake Sharif LTD
185.116.23.0	185.116.23.255	Tarahan Shabake Sharif LTD
185.116.24.0	185.116.27.255	Mashhad University of Medical Sciences
185.116.44.0	185.116.47.255	Iran Telecommunication Company Pjs
185.116.112.0	185.116.112.255	Arvin Data Transfer Co. Ltd.
185.116.160.0	185.116.160.255	Green Web Samaneh Novin PJSC
185.116.161.0	185.116.161.255	Green Web Samaneh Novin PJSC
185.116.162.0	185.116.162.255	Green Web Samaneh Novin PJSC
185.116.163.0	185.116.163.255	Green Web Samaneh Novin PJSC
185.117.48.0	185.117.51.255	Iran Telecommunication Company Pjs
185.117.136.0	185.117.139.255	Fara Ertebat Mabna PJS
185.117.204.0	185.117.204.255	Securities and Exchange Organization
185.117.205.0	185.117.205.255	Securities and Exchange Organization
185.117.206.0	185.117.206.255	Securities and Exchange Organization
185.117.207.0	185.117.207.255	Securities and Exchange Organization
185.118.12.0	185.118.12.255	Toos-Ashena PJSC
185.118.13.0	185.118.13.255	Cooperative Afra ertebatat-e-sabet-e Rasa Co
185.118.14.0	185.118.14.255	Cooperative Afra ertebatat-e-sabet-e Rasa Co
185.118.15.0	185.118.15.255	Toos-Ashena PJSC
185.118.136.0	185.118.139.255	Samen Ertebat Asr Co. (P.J.S.)
185.118.152.0	185.118.152.255	SINA ATINET Company (Ltd)
185.118.153.0	185.118.153.255	SINA ATINET Company (Ltd)
185.118.154.0	185.118.154.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
185.118.155.0	185.118.155.255	SINA ATINET Company (Ltd)
185.119.4.0	185.119.4.255	Toseh Samanehay-E Narmafzari-E Negin (Tosan) PJSC
185.119.5.0	185.119.5.255	Toseh Samanehay-E Narmafzari-E Negin (Tosan) PJSC
185.119.6.0	185.119.6.255	Toseh Samanehay-E Narmafzari-E Negin (Tosan) PJSC
185.119.7.0	185.119.7.255	Toseh Samanehay-E Narmafzari-E Negin (Tosan) PJSC
185.119.164.0	185.119.164.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
185.119.165.0	185.119.165.255	SINA ATINET Company (Ltd)
185.119.166.0	185.119.166.255	SINA ATINET Company (Ltd)
185.119.167.0	185.119.167.255	SINA ATINET Company (Ltd)
185.119.199.0	185.119.199.255	Fanavaran Ideh Pardaz Imenaria
185.119.240.0	185.119.243.255	Abramad Technological Infrastructures Development Company PJS
185.120.120.0	185.120.123.255	Ali Niknam
185.120.160.0	185.120.160.255	Khallagh Borhan Market Development for Creative Industries Co
185.120.161.0	185.120.161.255	Khallagh Borhan Market Development for Creative Industries Co
185.120.162.0	185.120.162.255	Khallagh Borhan Market Development for Creative Industries Co
185.120.163.0	185.120.163.255	Khallagh Borhan Market Development for Creative Industries Co
185.120.168.0	185.120.168.255	Dadeh Gostar Asr Novin P.J.S. Co.
185.120.169.0	185.120.169.255	Dadeh Gostar Asr Novin P.J.S. Co.
185.120.170.0	185.120.170.255	Rayaneh Danesh Golestan Complex P.J.S. Co.
185.120.171.0	185.120.171.255	Tejarat Iranian Technology Infrastructure PJSC
185.120.192.0	185.120.195.255	Pishgaman Toseeh Fanavari Etelaat Va Ertebatat Jonoub (Joint Stock Company)
185.120.196.0	185.120.199.255	Pishgaman Toseeh Fanavari Etelaat Va Ertebatat Jonoub (Joint Stock Company)
185.120.200.0	185.120.200.255	Pishgaman Toseeh Fanavari Etelaat Va Ertebatat Jonoub (Joint Stock Company)
185.120.201.0	185.120.201.255	Pishgaman Toseeh Fanavari Etelaat Va Ertebatat Jonoub (Joint Stock Company)
185.120.202.0	185.120.202.255	Pishgaman Toseeh Fanavari Etelaat Va Ertebatat Jonoub (Joint Stock Company)
185.120.203.0	185.120.203.255	Pishgaman Toseeh Fanavari Etelaat Va Ertebatat Jonoub (Joint Stock Company)
185.120.208.0	185.120.211.255	Pishgaman Toseeh Fanavari Etelaat Va Ertebatat Jonoub (Joint Stock Company)
185.120.212.0	185.120.215.255	Iran Telecommunication Company Pjs
185.120.216.0	185.120.219.255	Pishgaman Toseeh Fanavari Etelaat Va Ertebatat Jonoub (Joint Stock Company)
185.120.220.0	185.120.220.255	Asiatech Data Transmission Company
185.120.221.0	185.120.221.255	Asiatech Data Transmission Company
185.120.222.0	185.120.222.255	Asiatech Data Transmission Company
185.120.223.0	185.120.223.255	Asiatech Data Transmission Company
185.120.224.0	185.120.227.255	Pishgaman Toseeh Fanavari Etelaat Va Ertebatat Jonoub (Joint Stock Company)
185.120.228.0	185.120.231.255	Iran Telecommunication Company Pjs
185.120.232.0	185.120.235.255	Pishgaman Toseeh Fanavari Etelaat Va Ertebatat Jonoub (Joint Stock Company)
185.120.236.0	185.120.239.255	Iran Telecommunication Company Pjs
185.120.240.0	185.120.243.255	Pishgaman Toseeh Fanavari Etelaat Va Ertebatat Jonoub (Joint Stock Company)
185.120.244.0	185.120.245.255	Iran Telecommunication Company Pjs
185.120.246.0	185.120.247.255	Iran Telecommunication Company Pjs
185.120.248.0	185.120.249.255	Asiatech Data Transmission Company
185.120.250.0	185.120.251.255	Asiatech Data Transmission Company
185.121.56.0	185.121.56.255	Hesabgar Pardaz Gharb PJSC
185.121.57.0	185.121.57.255	Hesabgar Pardaz Gharb PJSC
185.121.58.0	185.121.58.255	Hesabgar Pardaz Gharb PJSC
185.121.59.0	185.121.59.255	Hesabgar Pardaz Gharb PJSC
185.121.128.0	185.121.128.255	Sefroyek Pardaz Engineering PJSC
185.121.129.0	185.121.129.255	Sefroyek Pardaz Engineering PJSC
185.121.130.0	185.121.130.255	Sefroyek Pardaz Engineering PJSC
185.121.131.0	185.121.131.255	Sefroyek Pardaz Engineering PJSC
185.122.80.0	185.122.80.255	Melli Pakhsh Faravardehay Nafti Iran PJSC
185.122.81.0	185.122.81.255	Melli Pakhsh Faravardehay Nafti Iran PJSC
185.122.82.0	185.122.82.255	Melli Pakhsh Faravardehay Nafti Iran PJSC
185.122.83.0	185.122.83.255	Melli Pakhsh Faravardehay Nafti Iran PJSC
185.123.68.0	185.123.71.255	Shahrood University of Technology
185.123.208.0	185.123.208.255	Ministry of Health Treatment and Medical Education
185.123.209.0	185.123.209.255	Ministry of Health Treatment and Medical Education
185.123.210.0	185.123.210.255	Ministry of Health Treatment and Medical Education
185.123.211.0	185.123.211.255	Ministry of Health Treatment and Medical Education
185.124.112.0	185.124.115.255	Rayankadeh Apadana Company Ltd
185.124.156.0	185.124.159.255	Iran Telecommunication Company Pjs
185.124.172.0	185.124.172.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
185.124.173.0	185.124.173.255	DyarWeb Information Technology Ltd
185.124.174.0	185.124.174.255	DyarWeb Information Technology Ltd
185.124.175.0	185.124.175.255	Ali Monfared
185.125.20.0	185.125.20.255	Hesabgar Pardaz Gharb PJSC
185.125.21.0	185.125.21.255	Hesabgar Pardaz Gharb PJSC
185.125.22.0	185.125.22.255	Hesabgar Pardaz Gharb PJSC
185.125.23.0	185.125.23.255	Hesabgar Pardaz Gharb PJSC
185.125.244.0	185.125.247.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
185.125.248.0	185.125.251.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
185.125.252.0	185.125.253.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
185.125.254.0	185.125.254.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
185.125.255.0	185.125.255.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
185.126.0.0	185.126.0.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
185.126.1.0	185.126.1.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
185.126.2.0	185.126.2.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
185.126.3.0	185.126.3.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
185.126.4.0	185.126.4.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
185.126.5.0	185.126.5.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
185.126.6.0	185.126.6.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
185.126.7.0	185.126.7.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
185.126.8.0	185.126.8.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
185.126.9.0	185.126.9.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
185.126.10.0	185.126.10.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
185.126.11.0	185.126.11.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
185.126.12.0	185.126.12.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
185.126.13.0	185.126.13.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
185.126.14.0	185.126.14.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
185.126.15.0	185.126.15.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
185.126.16.0	185.126.19.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
185.126.40.0	185.126.43.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
185.126.132.0	185.126.133.255	Abr Tose'eh Darya Group Company PJSC
185.126.200.0	185.126.200.255	Toesegaran Shabakeh Arseh Novin Ltd
185.126.201.0	185.126.201.255	Toesegaran Shabakeh Arseh Novin Ltd
185.126.202.0	185.126.202.255	Dade Pardazi Mobinhost Co Ltd
185.126.203.0	185.126.203.255	Dade Pardazi Mobinhost Co Ltd
185.128.40.0	185.128.40.255	Sindad Network Technology PJSC
185.128.48.0	185.128.51.255	Iranian Net Communication and Electronic Services Co PJS
185.128.80.0	185.128.80.255	Pars Databan Co
185.128.81.0	185.128.81.255	Pars Databan Co
185.128.82.0	185.128.82.255	Pars Databan Co
185.128.83.0	185.128.83.255	Pars Databan Co
185.128.136.0	185.128.136.255	Sefroyek Pardaz Engineering PJSC
185.128.137.0	185.128.137.255	Sefroyek Pardaz Engineering PJSC
185.128.138.0	185.128.138.255	Sefroyek Pardaz Engineering PJSC
185.128.139.0	185.128.139.255	Sefroyek Pardaz Engineering PJSC
185.128.152.0	185.128.152.255	Shabakeh Gostar Shahriyar Co. (Ltd.)
185.128.153.0	185.128.153.255	Shabakeh Gostar Shahriyar Co. (Ltd.)
185.128.154.0	185.128.155.255	Honar Rayaneh Pooya Andisheh PJSC
185.128.164.0	185.128.165.255	Honar Rayaneh Pooya Andisheh PJSC
185.128.166.0	185.128.166.255	Honar Rayaneh Pooya Andisheh PJSC
185.128.167.0	185.128.167.255	Asiatech Data Transmission Company
185.129.80.0	185.129.80.255	Paya Kian Parham Co. (PJS)
185.129.81.0	185.129.81.255	Paya Kian Parham Co. (PJS)
185.129.82.0	185.129.82.255	Paya Kian Parham Co. (PJS)
185.129.83.0	185.129.83.255	Paya Kian Parham Co. (PJS)
185.129.108.0	185.129.111.255	Pooyesh Axe Information and Communication Technology PJS
185.129.116.0	185.129.116.255	Asiatech Data Transmission Company
185.129.117.0	185.129.117.255	Asiatech Data Transmission Company
185.129.118.0	185.129.118.255	Asiatech Data Transmission Company
185.129.119.0	185.129.119.255	Asiatech Data Transmission Company
185.129.168.0	185.129.168.255	Abzarhaye Farsi Shabakeh Co. LTD.
185.129.169.0	185.129.169.255	Abzarhaye Farsi Shabakeh Co. LTD.
185.129.170.0	185.129.170.255	Abzarhaye Farsi Shabakeh Co. LTD.
185.129.171.0	185.129.171.255	Abzarhaye Farsi Shabakeh Co. LTD.
185.129.184.0	185.129.187.255	Pishgaman Tejarat Sayar Company (Private Joint Stock)
185.129.188.0	185.129.188.255	Pishgaman Tejarat Sayar Company (Private Joint Stock)
185.129.189.0	185.129.189.255	Shaparak Electronic Card Payment Network Company PJS
185.129.190.0	185.129.191.255	Pishgaman Tejarat Sayar Company (Private Joint Stock)
185.129.196.0	185.129.197.255	Gostaresh Ertebat Azin Kia Company PJSC
185.129.198.0	185.129.199.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
185.129.200.0	185.129.201.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
185.129.202.0	185.129.203.255	Gostaresh Ertebat Azin Kia Company PJSC
185.129.212.0	185.129.215.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
185.129.216.0	185.129.217.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
185.129.218.0	185.129.219.255	Gostaresh Ertebat Azin Kia Company PJSC
185.129.228.0	185.129.231.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
185.129.232.0	185.129.235.255	Enteghal Dadeh Mahan Co. PJSC
185.129.236.0	185.129.237.255	Enteghal Dadeh Mahan Co. PJSC
185.129.238.0	185.129.239.255	Enteghal Dadeh Mahan Co. PJSC
185.129.240.0	185.129.243.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
185.130.50.0	185.130.50.255	Mizban Dade Shetaban Co. (Ltd)
185.130.76.0	185.130.76.255	ICT Research Institute (Iran Telecommunications Research Center Itrc)
185.130.77.0	185.130.77.255	ICT Research Institute (Iran Telecommunications Research Center Itrc)
185.130.78.0	185.130.78.255	ICT Research Institute (Iran Telecommunications Research Center Itrc)
185.130.79.0	185.130.79.255	ICT Research Institute (Iran Telecommunications Research Center Itrc)
185.130.101.0	185.130.101.255	Information Technology Organization of Isfahan Municipality
185.131.28.0	185.131.29.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
185.131.30.0	185.131.31.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
185.131.84.0	185.131.87.255	Iran Telecommunication Company Pjs
185.131.88.0	185.131.91.255	Iran Telecommunication Company Pjs
185.131.92.0	185.131.95.255	Iran Telecommunication Company Pjs
185.131.100.0	185.131.103.255	Iran Telecommunication Company Pjs
185.131.108.0	185.131.111.255	Iran Telecommunication Company Pjs
185.131.112.0	185.131.115.255	Iran Telecommunication Company Pjs
185.131.116.0	185.131.119.255	Iran Telecommunication Company Pjs
185.131.124.0	185.131.127.255	Iran Telecommunication Company Pjs
185.131.128.0	185.131.131.255	Iran Telecommunication Company Pjs
185.131.136.0	185.131.139.255	Iran Telecommunication Company Pjs
185.131.140.0	185.131.143.255	Iran Telecommunication Company Pjs
185.131.148.0	185.131.151.255	Iran Telecommunication Company Pjs
185.131.152.0	185.131.155.255	Iran Telecommunication Company Pjs
185.131.156.0	185.131.159.255	Iran Telecommunication Company Pjs
185.131.164.0	185.131.167.255	Iran Telecommunication Company Pjs
185.131.168.0	185.131.169.255	Iran Telecommunication Company Pjs
185.131.170.0	185.131.171.255	Iran Telecommunication Company Pjs
185.132.80.0	185.132.81.255	Asiatech Data Transmission Company
185.132.82.0	185.132.83.255	Asiatech Data Transmission Company
185.132.124.0	185.132.124.255	Beh Pardakht Melat Co PJS
185.132.212.0	185.132.212.255	Eghtesad Novin Bank PJSC
185.132.213.0	185.132.213.255	Eghtesad Novin Bank PJSC
185.132.214.0	185.132.214.255	Eghtesad Novin Bank PJSC
185.132.215.0	185.132.215.255	Eghtesad Novin Bank PJSC
185.133.125.0	185.133.125.255	Web Dadeh Paydar Co (Ltd)
185.133.152.0	185.133.155.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
185.133.164.0	185.133.167.255	Iran Telecommunication Company Pjs
185.133.244.0	185.133.244.255	Tehran Securities Exchange Technology Management Co. PLC
185.133.245.0	185.133.245.255	Tehran Securities Exchange Technology Management Co. PLC
185.133.246.0	185.133.246.255	Tehran Securities Exchange Technology Management Co. PLC
185.133.247.0	185.133.247.255	Tehran Securities Exchange Technology Management Co. PLC
185.134.96.0	185.134.99.255	Yaghoot Pars Asia Cooperative Company
185.135.28.0	185.135.28.255	Sadad Informatics Corporation PJSC
185.135.29.0	185.135.29.255	Sadad Informatics Corporation PJSC
185.135.30.0	185.135.30.255	Sadad Informatics Corporation PJSC
185.135.31.0	185.135.31.255	Sadad Informatics Corporation PJSC
185.135.46.0	185.135.46.255	Parsian Hamrah Lotus Co.(Private Joint Stock)
185.135.47.0	185.135.47.255	Hamed Manavi Haghighi
185.135.228.0	185.135.229.255	Mobin Net Communication Company (Private Joint Stock)
185.135.230.0	185.135.230.255	Mobin Net Communication Company (Private Joint Stock)
185.135.231.0	185.135.231.255	Mobin Net Communication Company (Private Joint Stock)
185.136.100.0	185.136.103.255	Mobin Net Communication Company (Private Joint Stock)
185.136.133.0	185.136.133.255	Dade Samane Fanava Company (PJS)
185.136.135.0	185.136.135.255	DyarWeb Information Technology Ltd
185.136.172.0	185.136.175.255	Iran Telecommunication Company Pjs
185.136.180.0	185.136.180.255	Araax Dadeh Gostar Information and Communication Development Co (Private Joint Stock)
185.136.181.0	185.136.181.255	Araax Dadeh Gostar Information and Communication Development Co (Private Joint Stock)
185.136.182.0	185.136.182.255	Araax Dadeh Gostar Information and Communication Development Co (Private Joint Stock)
185.136.183.0	185.136.183.255	Araax Dadeh Gostar Information and Communication Development Co (Private Joint Stock)
185.136.192.0	185.136.195.255	Honar Rayaneh Pooya Andisheh PJSC
185.136.220.0	185.136.223.255	Iran Telecommunication Company Pjs
185.137.24.0	185.137.24.255	Rasaneh Avabarid Private Joint Stock Company
185.137.25.0	185.137.25.255	Rasaneh Avabarid Private Joint Stock Company
185.137.26.0	185.137.26.255	Rasaneh Avabarid Private Joint Stock Company
185.137.27.0	185.137.27.255	Rasaneh Avabarid Private Joint Stock Company
185.137.60.0	185.137.60.255	Hesabgar Pardaz Gharb PJSC
185.137.61.0	185.137.61.255	Hesabgar Pardaz Gharb PJSC
185.137.62.0	185.137.62.255	Hesabgar Pardaz Gharb PJSC
185.137.63.0	185.137.63.255	Hesabgar Pardaz Gharb PJSC
185.137.108.0	185.137.108.255	Fara Negar Pardaz Noor Khuzestan Co.JSP
185.137.109.0	185.137.109.255	Fara Negar Pardaz Noor Khuzestan Co.JSP
185.137.110.0	185.137.110.255	Fara Negar Pardaz Noor Khuzestan Co.JSP
185.137.111.0	185.137.111.255	Fara Negar Pardaz Noor Khuzestan Co.JSP
185.139.64.0	185.139.65.255	Medical Science University of Kashan
185.139.66.0	185.139.67.255	Medical Science University of Kashan
185.140.4.0	185.140.4.255	Rasaneh Mehr Vatan Co. PJS
185.140.5.0	185.140.5.255	Rasaneh Mehr Vatan Co. PJS
185.140.6.0	185.140.6.255	Rasaneh Mehr Vatan Co. PJS
185.140.7.0	185.140.7.255	Rasaneh Mehr Vatan Co. PJS
185.140.56.0	185.140.59.255	Rayankadeh Apadana Company Ltd
185.140.232.0	185.140.232.255	Pooya Parto Qeshm Cooperative Company
185.140.233.0	185.140.233.255	Pooya Parto Qeshm Cooperative Company
185.140.234.0	185.140.234.255	Pooya Parto Qeshm Cooperative Company
185.140.235.0	185.140.235.255	Pooya Parto Qeshm Cooperative Company
185.141.36.0	185.141.39.255	Cooperative Afra ertebatat-e-sabet-e Rasa Co
185.141.48.0	185.141.48.255	Islamic Azad University
185.141.49.0	185.141.49.255	Islamic Azad University
185.141.50.0	185.141.50.255	Islamic Azad University
185.141.51.0	185.141.51.255	Islamic Azad University
185.141.104.0	185.141.104.255	Sefroyek Pardaz Engineering PJSC
185.141.105.0	185.141.105.255	Sefroyek Pardaz Engineering PJSC
185.141.106.0	185.141.106.255	Sefroyek Pardaz Engineering PJSC
185.141.107.0	185.141.107.255	Sefroyek Pardaz Engineering PJSC
185.141.132.0	185.141.132.255	Sefroyek Pardaz Engineering PJSC
185.141.133.0	185.141.133.255	Sefroyek Pardaz Engineering PJSC
185.141.134.0	185.141.134.255	Sefroyek Pardaz Engineering PJSC
185.141.135.0	185.141.135.255	Sefroyek Pardaz Engineering PJSC
185.141.168.0	185.141.168.255	Asiatech Data Transmission Company
185.141.169.0	185.141.169.255	Asiatech Data Transmission Company
185.141.170.0	185.141.170.255	Asiatech Data Transmission Company
185.141.171.0	185.141.171.255	Asiatech Data Transmission Company
185.141.212.0	185.141.212.255	Asiatech Data Transmission Company
185.141.213.0	185.141.213.255	Asiatech Data Transmission Company
185.141.214.0	185.141.214.255	Asiatech Data Transmission Company
185.141.215.0	185.141.215.255	Asiatech Data Transmission Company
185.141.244.0	185.141.244.255	Iman Samaneh Sepehr LLC
185.141.245.0	185.141.245.255	Iman Samaneh Sepehr LLC
185.141.246.0	185.141.246.255	Iman Samaneh Sepehr LLC
185.141.247.0	185.141.247.255	Iman Samaneh Sepehr LLC
185.142.92.0	185.142.92.255	Hesabgar Pardaz Gharb PJSC
185.142.93.0	185.142.93.255	Hesabgar Pardaz Gharb PJSC
185.142.94.0	185.142.94.255	Hesabgar Pardaz Gharb PJSC
185.142.95.0	185.142.95.255	Hesabgar Pardaz Gharb PJSC
185.142.124.0	185.142.124.255	Hesabgar Pardaz Gharb PJSC
185.142.125.0	185.142.125.255	Hesabgar Pardaz Gharb PJSC
185.142.126.0	185.142.126.255	Hesabgar Pardaz Gharb PJSC
185.142.127.0	185.142.127.255	Hesabgar Pardaz Gharb PJSC
185.142.156.0	185.142.156.255	Asre Pardazeshe Ettelaate Amin Institute
185.142.157.0	185.142.157.255	Asre Pardazeshe Ettelaate Amin Institute
185.142.158.0	185.142.158.255	Asre Pardazeshe Ettelaate Amin Institute
185.142.159.0	185.142.159.255	Asre Pardazeshe Ettelaate Amin Institute
185.142.232.0	185.142.232.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
185.142.233.0	185.142.233.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
185.142.234.0	185.142.235.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
185.143.73.0	185.143.73.255	Farhang Azma Communications Company LTD
185.143.204.0	185.143.207.255	Rayankadeh Apadana Company Ltd
185.143.232.0	185.143.232.255	Noyan Abr Arvan Co. ( Private Joint Stock)
185.143.233.0	185.143.233.255	Noyan Abr Arvan Co. ( Private Joint Stock)
185.143.234.0	185.143.234.255	Noyan Abr Arvan Co. ( Private Joint Stock)
185.143.235.0	185.143.235.255	Noyan Abr Arvan Co. ( Private Joint Stock)
185.144.64.0	185.144.64.255	Kerman University of Medical Sciences
185.144.65.0	185.144.65.255	Kerman University of Medical Sciences
185.144.66.0	185.144.66.255	Kerman University of Medical Sciences
185.144.67.0	185.144.67.255	Kerman University of Medical Sciences
185.145.8.0	185.145.11.255	Rayankadeh Apadana Company Ltd
185.145.184.0	185.145.187.255	Pooya Parto Qeshm Cooperative Company
185.147.40.0	185.147.43.255	Iran Telecommunication Company Pjs
185.147.84.0	185.147.84.255	Ports and Martime Organization
185.147.85.0	185.147.85.255	Ports and Martime Organization
185.147.86.0	185.147.86.255	Ports and Martime Organization
185.147.87.0	185.147.87.255	Ports and Martime Organization
185.147.160.0	185.147.160.255	Asiatech Data Transmission Company
185.147.161.0	185.147.161.255	Asiatech Data Transmission Company
185.147.162.0	185.147.162.255	Asiatech Data Transmission Company
185.147.163.0	185.147.163.255	Asiatech Data Transmission Company
185.147.176.0	185.147.176.255	Fannavaran-e Idea Pardaz-e Saba PJSC
185.147.177.0	185.147.177.255	Fannavaran-e Idea Pardaz-e Saba PJSC
185.147.178.0	185.147.178.255	Fannavaran-e Idea Pardaz-e Saba PJSC
185.147.179.0	185.147.179.255	Fannavaran-e Idea Pardaz-e Saba PJSC
185.149.192.0	185.149.192.255	Sindad Network Technology PJSC
185.150.108.0	185.150.111.255	Rayanmehr Danesh Sanj Company Ltd
185.151.236.0	185.151.239.255	Pooyesh Axe Information and Communication Technology PJS
185.153.184.0	185.153.184.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
185.153.185.0	185.153.185.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
185.153.186.0	185.153.187.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
185.153.208.0	185.153.211.255	Persia System Kashan Ltd
185.154.184.0	185.154.187.255	Khuzestan Water and Power Authority Co. PJS
185.154.190.0	185.154.190.255	Atis Omran Sevin PSJ
185.155.8.0	185.155.9.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
185.155.10.0	185.155.10.255	Gostaresh Ertebat Azin Kia Company PJSC
185.155.11.0	185.155.11.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
185.155.12.0	185.155.13.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
185.155.14.0	185.155.15.255	Gostaresh Ertebat Azin Kia Company PJSC
185.155.72.0	185.155.72.255	Web Dadeh Paydar Co (Ltd)
185.155.73.0	185.155.73.255	Seyed Amir Tabatabaei Anaraki
185.155.236.0	185.155.236.255	Communication Regulatory Authority
185.155.237.0	185.155.237.255	Communication Regulatory Authority
185.155.238.0	185.155.238.255	Communication Regulatory Authority
185.155.239.0	185.155.239.255	Communication Regulatory Authority
185.157.8.0	185.157.9.255	Pars Oil & Gas Company (Private Joint Stock)
185.157.10.0	185.157.11.255	Pars Oil & Gas Company (Private Joint Stock)
185.158.172.0	185.158.172.255	Cooperative Afra ertebatat-e-sabet-e Rasa Co
185.158.173.0	185.158.173.255	Cooperative Afra ertebatat-e-sabet-e Rasa Co
185.158.174.0	185.158.174.255	Cooperative Afra ertebatat-e-sabet-e Rasa Co
185.158.175.0	185.158.175.255	Tahlil Dadeh Novin Fadak LLC
185.159.152.0	185.159.152.255	Fanavari Serverpars Argham Gostar Company Ltd.
185.159.153.0	185.159.153.255	Fanavari Serverpars Argham Gostar Company Ltd.
185.159.154.0	185.159.154.255	Fanavari Serverpars Argham Gostar Company Ltd.
185.159.155.0	185.159.155.255	Fanavari Serverpars Argham Gostar Company Ltd.
185.159.176.0	185.159.179.255	Pardis Etelaresan Sepehr Employee's Cooperative Co.
185.159.189.0	185.159.189.255	Sindad Network Technology PJSC
185.160.104.0	185.160.104.255	Kharazmi University
185.160.105.0	185.160.105.255	Kharazmi University
185.160.106.0	185.160.106.255	Kharazmi University
185.160.107.0	185.160.107.255	Kharazmi University
185.160.176.0	185.160.176.255	Atieh Dadeh Pardaz PJSC
185.160.177.0	185.160.177.255	Atieh Dadeh Pardaz PJSC
185.160.178.0	185.160.178.255	Atieh Dadeh Pardaz PJSC
185.160.179.0	185.160.179.255	Atieh Dadeh Pardaz PJSC
185.160.205.0	185.160.205.255	Negin Narmafzar Asak Research and Information Cooperative Company
185.161.36.0	185.161.36.255	Atrin Information & Communications Technology Company PJS
185.161.37.0	185.161.37.255	Atrin Information & Communications Technology Company PJS
185.161.38.0	185.161.38.255	Atrin Information & Communications Technology Company PJS
185.161.39.0	185.161.39.255	Atrin Information & Communications Technology Company PJS
185.161.112.0	185.161.112.255	Parvaz System Information Technology Company (LTD)
185.161.113.0	185.161.113.255	Parvaz System Information Technology Company (LTD)
185.161.114.0	185.161.114.255	Parvaz System Information Technology Company (LTD)
185.161.115.0	185.161.115.255	Parvaz System Information Technology Company (LTD)
185.161.121.0	185.161.121.255	Sindad Network Technology PJSC
185.161.250.0	185.161.250.255	Pardazesh Abri Risman Ltd
185.162.40.0	185.162.43.255	Mobin Net Communication Company (Private Joint Stock)
185.162.216.0	185.162.216.255	Pars Fonoun Ofogh Information Technology and Communications Company Ltd
185.162.217.0	185.162.217.255	Asiatech Data Transmission Company
185.162.218.0	185.162.218.255	Asiatech Data Transmission Company
185.162.219.0	185.162.219.255	Asiatech Data Transmission Company
185.164.72.0	185.164.72.255	Vandad Vira Hooman LLC
185.164.73.0	185.164.73.255	Vandad Vira Hooman LLC
185.164.74.0	185.164.74.255	Atrin Information & Communications Technology Company PJS
185.164.75.0	185.164.75.255	Atrin Information & Communications Technology Company PJS
185.164.252.0	185.164.252.255	Ertebat Gostare Nakhl Jonoob Company PJSC
185.164.253.0	185.164.253.255	Ertebat Gostare Nakhl Jonoob Company PJSC
185.164.254.0	185.164.254.255	Atrin Information & Communications Technology Company PJS
185.164.255.0	185.164.255.255	Atrin Information & Communications Technology Company PJS
185.165.28.0	185.165.28.255	Mizban Amvaj Sahel Sepehr Bushehr PJSC
185.165.29.0	185.165.29.255	Noavaran Shabakeh Sabz Mehregan (Ltd.)
185.165.30.0	185.165.30.255	Mizban Amvaj Sahel Sepehr Bushehr PJSC
185.165.31.0	185.165.31.255	Mizban Amvaj Sahel Sepehr Bushehr PJSC
185.165.40.0	185.165.43.255	Mizban Web Paytakht Co. Ltd.
185.165.116.0	185.165.116.255	Dadeh Gostar Parmis PJS Company
185.165.117.0	185.165.117.255	Dadeh Gostar Parmis PJS Company
185.165.118.0	185.165.118.255	Dadeh Gostar Parmis PJS Company
185.165.119.0	185.165.119.255	Dadeh Gostar Parmis PJS Company
185.165.204.0	185.165.204.255	Sima Rayan Sharif Company (Ltd.)
185.165.205.0	185.165.205.255	Sima Rayan Sharif Company (Ltd.)
185.165.206.0	185.165.206.255	Sima Rayan Sharif Company (Ltd.)
185.165.207.0	185.165.207.255	Iran Telecommunication Company Pjs
185.166.60.0	185.166.63.255	SINA ATINET Company (Ltd)
185.166.104.0	185.166.104.255	Hezardastan Unit Cloud Computing PJSC
185.166.105.0	185.166.105.255	Hezardastan Unit Cloud Computing PJSC
185.166.106.0	185.166.106.255	Hezardastan Unit Cloud Computing PJSC
185.166.107.0	185.166.107.255	Hezardastan Unit Cloud Computing PJSC
185.166.112.0	185.166.112.255	Tarashe Sabz Tehran Co. Ltd.
185.166.113.0	185.166.113.255	Aria Shatel PJSC
185.166.114.0	185.166.115.255	Tarashe Sabz Tehran Co. Ltd.
185.167.72.0	185.167.72.255	Shaparak Electronic Card Payment Network Company PJS
185.167.73.0	185.167.73.255	Shaparak Electronic Card Payment Network Company PJS
185.167.74.0	185.167.74.255	Shaparak Electronic Card Payment Network Company PJS
185.167.75.0	185.167.75.255	Shaparak Electronic Card Payment Network Company PJS
185.167.124.0	185.167.124.255	Ati Sazan Farhang Tasnim Intitute
185.167.125.0	185.167.125.255	Ati Sazan Farhang Tasnim Intitute
185.167.126.0	185.167.126.255	Ati Sazan Farhang Tasnim Intitute
185.167.127.0	185.167.127.255	Ati Sazan Farhang Tasnim Intitute
185.169.6.0	185.169.6.255	Noavaran Shabakeh Sabz Mehregan (Ltd.)
185.169.20.0	185.169.23.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
185.169.36.0	185.169.39.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
185.170.8.0	185.170.8.255	Sefroyek Pardaz Engineering PJSC
185.170.236.0	185.170.236.255	Cooperative Afra ertebatat-e-sabet-e Rasa Co
185.170.237.0	185.170.237.255	Cooperative Afra ertebatat-e-sabet-e Rasa Co
185.170.238.0	185.170.238.255	Cooperative Afra ertebatat-e-sabet-e Rasa Co
185.170.239.0	185.170.239.255	Cooperative Afra ertebatat-e-sabet-e Rasa Co
185.171.52.0	185.171.52.255	Aranik Communications Technology PJSC
185.171.53.0	185.171.53.255	Rasaneh Avabarid Private Joint Stock Company
185.171.54.0	185.171.54.255	Sefroyek Pardaz Engineering PJSC
185.171.55.0	185.171.55.255	Sefroyek Pardaz Engineering PJSC
185.172.0.0	185.172.3.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
185.172.68.0	185.172.69.255	Rayan Ravesh Sena Ltd
185.172.70.0	185.172.71.255	Rayan Ravesh Sena Ltd
185.172.212.0	185.172.212.255	Lesun Communication Furtherance Engineers Co (Ltd.)
185.172.213.0	185.172.213.255	Lesun Communication Furtherance Engineers Co (Ltd.)
185.172.214.0	185.172.214.255	Lesun Communication Furtherance Engineers Co (Ltd.)
185.172.215.0	185.172.215.255	Lesun Communication Furtherance Engineers Co (Ltd.)
185.173.104.0	185.173.104.255	Noavaran Shabakeh Sabz Mehregan (Ltd.)
185.173.105.0	185.173.105.255	Noavaran Shabakeh Sabz Mehregan (Ltd.)
185.173.106.0	185.173.106.255	Noavaran Shabakeh Sabz Mehregan (Ltd.)
185.173.107.0	185.173.107.255	Noavaran Shabakeh Sabz Mehregan (Ltd.)
185.173.129.0	185.173.129.255	Sefroyek Pardaz Engineering PJSC
185.173.130.0	185.173.130.255	Sefroyek Pardaz Engineering PJSC
185.173.168.0	185.173.169.255	Fereidoon Hedayatinia
185.173.170.0	185.173.170.255	Fereidoon Hedayatinia
185.173.171.0	185.173.171.255	Fereidoon Hedayatinia
185.174.132.0	185.174.132.255	Cooperative Afra ertebatat-e-sabet-e Rasa Co
185.174.134.0	185.174.134.255	Cooperative Afra ertebatat-e-sabet-e Rasa Co
185.174.200.0	185.174.203.255	Fara Ertebat Mabna PJS
185.174.248.0	185.174.249.255	Khallagh Borhan Market Development for Creative Industries Co
185.174.250.0	185.174.250.255	Khallagh Borhan Market Development for Creative Industries Co
185.174.251.0	185.174.251.255	Khallagh Borhan Market Development for Creative Industries Co
185.175.76.0	185.175.77.255	Mobin Net Communication Company (Private Joint Stock)
185.175.78.0	185.175.79.255	Mobin Net Communication Company (Private Joint Stock)
185.175.240.0	185.175.243.255	Bistalk Telecumunication Systems Company (P.J.S.)
185.176.32.0	185.176.35.255	Mobin Net Communication Company (Private Joint Stock)
185.176.56.0	185.176.56.255	Homaye Jahan Nama Co. ( Private Joint Stock)
185.176.57.0	185.176.57.255	Homaye Jahan Nama Co. ( Private Joint Stock)
185.176.58.0	185.176.58.255	Homaye Jahan Nama Co. ( Private Joint Stock)
185.176.59.0	185.176.59.255	Homaye Jahan Nama Co. ( Private Joint Stock)
185.177.156.0	185.177.159.255	Fara Ertebat Mabna PJS
185.177.232.0	185.177.235.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
185.178.104.0	185.178.104.255	Iman Hamrah Behpardaz Company- Ltd
185.178.105.0	185.178.105.255	Iman Hamrah Behpardaz Company- Ltd
185.178.106.0	185.178.107.255	Iman Hamrah Behpardaz Company- Ltd
185.178.220.0	185.178.220.255	Information and Communication Technology Organization of Tabriz Municipality
185.178.221.0	185.178.221.255	Information and Communication Technology Organization of Tabriz Municipality
185.178.222.0	185.178.222.255	Information and Communication Technology Organization of Tabriz Municipality
185.178.223.0	185.178.223.255	Information and Communication Technology Organization of Tabriz Municipality
185.179.90.0	185.179.90.255	Faraso Samaneh Pasargad Ltd.
185.179.168.0	185.179.171.255	Enteghal Dadeh Mahan Co. PJSC
185.179.220.0	185.179.223.255	Fara Ertebat Mabna PJS
185.180.52.0	185.180.55.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
185.180.128.0	185.180.128.255	Azarakhsh Ava-E Ahvaz Co
185.180.129.0	185.180.129.255	Azarakhsh Ava-E Ahvaz Co
185.180.130.0	185.180.130.255	Azarakhsh Ava-E Ahvaz Co
185.180.131.0	185.180.131.255	Azarakhsh Ava-E Ahvaz Co
185.181.180.0	185.181.180.255	Computer Research Center of Islamic Sciences
185.181.181.0	185.181.181.255	Computer Research Center of Islamic Sciences
185.181.182.0	185.181.182.255	Computer Research Center of Islamic Sciences
185.181.183.0	185.181.183.255	Computer Research Center of Islamic Sciences
185.182.220.0	185.182.220.255	Mobin Net Communication Company (Private Joint Stock)
185.182.221.0	185.182.221.255	Mobin Net Communication Company (Private Joint Stock)
185.182.222.0	185.182.223.255	Mobin Net Communication Company (Private Joint Stock)
185.182.248.0	185.182.249.255	Sefroyek Pardaz Engineering PJSC
185.182.250.0	185.182.251.255	Sefroyek Pardaz Engineering PJSC
185.184.32.0	185.184.32.255	Novin Insurance Co. PJS
185.184.33.0	185.184.33.255	Novin Insurance Co. PJS
185.184.34.0	185.184.34.255	Novin Insurance Co. PJS
185.184.35.0	185.184.35.255	Novin Insurance Co. PJS
185.185.16.0	185.185.16.255	Seyed Amir Tabatabaei Anaraki
185.185.17.0	185.185.17.255	Seyed Amir Tabatabaei Anaraki
185.185.18.0	185.185.19.255	Seyed Amir Tabatabaei Anaraki
185.185.240.0	185.185.240.255	Secure Infrastructure of Transactional Services Company PJSC
185.185.241.0	185.185.241.255	Secure Infrastructure of Transactional Services Company PJSC
185.185.242.0	185.185.242.255	Secure Infrastructure of Transactional Services Company PJSC
185.185.243.0	185.185.243.255	Secure Infrastructure of Transactional Services Company PJSC
185.186.48.0	185.186.48.255	Pooya Parto Qeshm Cooperative Company
185.186.49.0	185.186.49.255	Pooya Parto Qeshm Cooperative Company
185.186.50.0	185.186.50.255	Pooya Parto Qeshm Cooperative Company
185.186.51.0	185.186.51.255	Pooya Parto Qeshm Cooperative Company
185.186.240.0	185.186.240.255	Imen Sanat Novin Alighapoo Private Joint Stock Company
185.186.241.0	185.186.241.255	Pouya Nazm Najafabad Computer Co(PJS)
185.186.242.0	185.186.242.255	Gostaresh Ertebat Azin Kia Company PJSC
185.186.243.0	185.186.243.255	Gostaresh Ertebat Azin Kia Company PJSC
185.187.48.0	185.187.48.255	Fara Net Gostar Kavoshgaran Aftab
185.187.49.0	185.187.49.255	Fara Net Gostar Kavoshgaran Aftab
185.187.50.0	185.187.50.255	Computer Research Center of Islamic Sciences
185.187.51.0	185.187.51.255	Web Gostaran Bandar Company (PJS)
185.187.84.0	185.187.84.255	Hamkaran System Co. PJS
185.187.85.0	185.187.85.255	Hamkaran System Co. PJS
185.187.86.0	185.187.86.255	Imen Sanat Novin Alighapoo Private Joint Stock Company
185.187.87.0	185.187.87.255	Imen Sanat Novin Alighapoo Private Joint Stock Company
185.188.104.0	185.188.104.255	Noavaran Fan Avazeh Co. PJS
185.188.105.0	185.188.105.255	Noavaran Fan Avazeh Co. PJS
185.188.106.0	185.188.106.255	Noavaran Fan Avazeh Co. PJS
185.188.107.0	185.188.107.255	Noavaran Fan Avazeh Co. PJS
185.188.113.0	185.188.113.255	Shabakieh Isfahan Co PJSC
185.188.114.0	185.188.114.255	Negin Ertebatate Ava Company PJS
185.188.115.0	185.188.115.255	Negin Ertebatate Ava Company PJS
185.189.120.0	185.189.123.255	Asiatech Data Transmission Company
185.190.20.0	185.190.23.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
185.190.39.0	185.190.39.255	Vandad Vira Hooman LLC
185.191.76.0	185.191.76.255	Rasaneh Avabarid Private Joint Stock Company
185.191.77.0	185.191.77.255	Rasaneh Avabarid Private Joint Stock Company
185.191.78.0	185.191.78.255	Rasaneh Avabarid Private Joint Stock Company
185.191.79.0	185.191.79.255	Rasaneh Avabarid Private Joint Stock Company
185.192.8.0	185.192.8.255	Bank Saderat Iran PJSC
185.192.9.0	185.192.9.255	Bank Saderat Iran PJSC
185.192.10.0	185.192.10.255	Bank Saderat Iran PJSC
185.192.11.0	185.192.11.255	Bank Saderat Iran PJSC
185.192.112.0	185.192.112.255	Parsan Lin Co. PJS
185.192.113.0	185.192.113.255	Parsan Lin Co. PJS
185.192.114.0	185.192.114.255	Parsan Lin Co. PJS
185.193.47.0	185.193.47.255	Tehran Stock Exchange Company (Public Joint Stock)
185.193.208.0	185.193.211.255	Khalij Fars Ettela Resan LTD
185.194.76.0	185.194.77.255	Khallagh Borhan Market Development for Creative Industries Co
185.194.78.0	185.194.78.255	Khallagh Borhan Market Development for Creative Industries Co
185.194.79.0	185.194.79.255	Khallagh Borhan Market Development for Creative Industries Co
185.194.244.0	185.194.244.255	IRIB (Islamic Republic of Iran Broadcasting)
185.194.245.0	185.194.245.255	IRIB (Islamic Republic of Iran Broadcasting)
185.194.246.0	185.194.246.255	IRIB (Islamic Republic of Iran Broadcasting)
185.194.247.0	185.194.247.255	IRIB (Islamic Republic of Iran Broadcasting)
185.195.72.0	185.195.75.255	Marun Petrochemical Co.
185.196.148.0	185.196.148.255	International Systems Engineering & Automation Company PJS
185.196.149.0	185.196.149.255	International Systems Engineering & Automation Company PJS
185.196.150.0	185.196.150.255	Mobarakeh Steel Company
185.196.151.0	185.196.151.255	South Hormozgan Steel Pjsc
185.197.68.0	185.197.71.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
185.197.112.0	185.197.115.255	University of Tehran
185.198.160.0	185.198.163.255	Iran Health Insurance Organization
185.199.208.0	185.199.208.255	Negin Ertebatate Ava Company PJS
185.199.210.0	185.199.210.255	Negin Ertebatate Ava Company PJS
185.199.211.0	185.199.211.255	Negin Ertebatate Ava Company PJS
185.201.48.0	185.201.48.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
185.201.49.0	185.201.49.255	Amir Hosein Maaref
185.201.50.0	185.201.50.255	Dadeh Pardaz Pouya-E Sharif LLC
185.201.51.0	185.201.51.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
185.202.56.0	185.202.56.255	Yaghoot Pars Asia Cooperative Company
185.202.57.0	185.202.57.255	Yaghoot Pars Asia Cooperative Company
185.202.58.0	185.202.58.255	Yaghoot Pars Asia Cooperative Company
185.202.59.0	185.202.59.255	Yaghoot Pars Asia Cooperative Company
185.203.160.0	185.203.160.255	Sadad Electronic Payment Company P.J.S.
185.203.161.0	185.203.161.255	Sadad Electronic Payment Company P.J.S.
185.203.162.0	185.203.162.255	Sadad Electronic Payment Company P.J.S.
185.203.163.0	185.203.163.255	Sadad Electronic Payment Company P.J.S.
185.204.180.0	185.204.180.255	Ertebatat Sabet Parsian Co. PJS
185.204.181.0	185.204.181.255	Ertebatat Sabet Parsian Co. PJS
185.204.182.0	185.204.183.255	Ertebatat Sabet Parsian Co. PJS
185.204.197.0	185.204.197.255	Vandad Vira Hooman LLC
185.205.203.0	185.205.203.255	Noavaran Shabakeh Sabz Mehregan (Ltd.)
185.206.92.0	185.206.95.255	Noyan Abr Arvan Co. ( Private Joint Stock)
185.206.229.0	185.206.229.255	Andisheh Sabz Mahan Co. Ltd
185.206.231.0	185.206.231.255	Sefroyek Pardaz Engineering PJSC
185.206.236.0	185.206.236.255	Ertebat Gostare Nakhl Jonoob Company PJSC
185.206.237.0	185.206.237.255	Ertebat Gostare Nakhl Jonoob Company PJSC
185.206.238.0	185.206.238.255	Ertebat Gostare Nakhl Jonoob Company PJSC
185.206.239.0	185.206.239.255	Ertebat Gostare Nakhl Jonoob Company PJSC
185.207.52.0	185.207.55.255	Iran Airports & Air Navigation Company (Private Joint Stock)
185.207.72.0	185.207.72.255	Central Insurance of Iran and Insurance Operations JSC
185.207.73.0	185.207.73.255	Central Insurance of Iran and Insurance Operations JSC
185.207.74.0	185.207.75.255	Central Insurance of Iran and Insurance Operations JSC
185.208.76.0	185.208.76.255	Noavaran Shabakeh Sabz Mehregan (Ltd.)
185.208.77.0	185.208.77.255	Noavaran Shabakeh Sabz Mehregan (Ltd.)
185.208.78.0	185.208.78.255	Noavaran Shabakeh Sabz Mehregan (Ltd.)
185.208.79.0	185.208.79.255	Noavaran Shabakeh Sabz Mehregan (Ltd.)
185.208.148.0	185.208.148.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
185.208.149.0	185.208.149.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
185.208.150.0	185.208.151.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
185.208.174.0	185.208.174.255	Vandad Vira Hooman LLC
185.208.175.0	185.208.175.255	Vandad Vira Hooman LLC
185.208.180.0	185.208.180.255	Asre Pardazeshe Ettelaate Amin Institute
185.208.181.0	185.208.181.255	Asre Pardazeshe Ettelaate Amin Institute
185.208.182.0	185.208.182.255	Asre Pardazeshe Ettelaate Amin Institute
185.208.183.0	185.208.183.255	Asre Pardazeshe Ettelaate Amin Institute
185.209.42.0	185.209.42.255	Toesegaran Shabakeh Arseh Novin Ltd
185.209.188.0	185.209.191.255	SINA ATINET Company (Ltd)
185.210.200.0	185.210.200.255	Asan Pardakht Persian Co. Ltd
185.210.201.0	185.210.201.255	Asan Pardakht Persian Co. Ltd
185.210.202.0	185.210.202.255	Asan Pardakht Persian Co. Ltd
185.210.203.0	185.210.203.255	Asan Pardakht Persian Co. Ltd
185.211.84.0	185.211.87.255	Pardis Fanvari Partak Ltd
185.211.88.0	185.211.88.255	Amirkabir University of Technology
185.211.89.0	185.211.89.255	Amirkabir University of Technology
185.211.90.0	185.211.90.255	Amirkabir University of Technology
185.211.91.0	185.211.91.255	Amirkabir University of Technology
185.212.48.0	185.212.48.255	Azarakhsh Ava-E Ahvaz Co
185.212.49.0	185.212.49.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
185.212.50.0	185.212.50.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
185.212.51.0	185.212.51.255	Enteghal Dade Arya Sarv Co (Ltd.)
185.212.192.0	185.212.192.255	Hesabgar Pardaz Gharb PJSC
185.212.193.0	185.212.193.255	Hesabgar Pardaz Gharb PJSC
185.212.194.0	185.212.194.255	Taknet Afzar Aftab PJSC
185.212.195.0	185.212.195.255	Hesabgar Pardaz Gharb PJSC
185.213.8.0	185.213.8.255	The Islamic Propagation Office of Qom Seminary
185.213.9.0	185.213.9.255	The Islamic Propagation Office of Qom Seminary
185.213.10.0	185.213.10.255	Asre Pardazeshe Ettelaate Amin Institute
185.213.11.0	185.213.11.255	The Islamic Propagation Office of Qom Seminary
185.213.164.0	185.213.164.255	Green Web Samaneh Novin PJSC
185.213.165.0	185.213.165.255	Green Web Samaneh Novin PJSC
185.213.166.0	185.213.166.255	Green Web Samaneh Novin PJSC
185.213.167.0	185.213.167.255	Green Web Samaneh Novin PJSC
185.213.195.0	185.213.195.255	Sefroyek Pardaz Engineering PJSC
185.214.36.0	185.214.36.255	Khalij Fars Ettela Resan LTD
185.214.37.0	185.214.37.255	Khalij Fars Ettela Resan LTD
185.214.38.0	185.214.38.255	Khalij Fars Ettela Resan LTD
185.214.39.0	185.214.39.255	Khalij Fars Ettela Resan LTD
185.215.124.0	185.215.124.255	Rahyab Payam Gostaran Co. P.J.S
185.215.125.0	185.215.125.255	Rahyab Payam Gostaran Co. P.J.S
185.215.126.0	185.215.126.255	Rahyab Payam Gostaran Co. P.J.S
185.215.127.0	185.215.127.255	Rahyab Payam Gostaran Co. P.J.S
185.215.152.0	185.215.153.255	Railways of Islamic Republic of Iran
185.215.154.0	185.215.154.255	Railways of Islamic Republic of Iran
185.215.155.0	185.215.155.255	Railways of Islamic Republic of Iran
185.215.228.0	185.215.228.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
185.215.229.0	185.215.229.255	Milad Ahadpour
185.215.230.0	185.215.230.255	DyarWeb Information Technology Ltd
185.215.231.0	185.215.231.255	Pars Abr Toseeh Ertebatat Ltd
185.217.6.0	185.217.6.255	Atrin Information & Communications Technology Company PJS
185.219.112.0	185.219.112.255	Ideh Gozin Ertebatat E Romak Co. Ltd
185.219.113.0	185.219.113.255	Ideh Gozin Ertebatat E Romak Co. Ltd
185.219.114.0	185.219.115.255	Ideh Gozin Ertebatat E Romak Co. Ltd
185.220.224.0	185.220.224.255	Noyan Abr Arvan Co. ( Private Joint Stock)
185.220.225.0	185.220.225.255	Noyan Abr Arvan Co. ( Private Joint Stock)
185.220.226.0	185.220.226.255	Noyan Abr Arvan Co. ( Private Joint Stock)
185.220.227.0	185.220.227.255	Noyan Abr Arvan Co. ( Private Joint Stock)
185.221.192.0	185.221.192.255	Hesabgar Pardaz Gharb PJSC
185.221.193.0	185.221.193.255	Taknet Afzar Aftab PJSC
185.221.194.0	185.221.194.255	Hesabgar Pardaz Gharb PJSC
185.221.195.0	185.221.195.255	Hesabgar Pardaz Gharb PJSC
185.221.239.0	185.221.239.255	Dade Samane Fanava Company (PJS)
185.222.120.0	185.222.120.255	Respina Networks & Beyond PJSC
185.222.121.0	185.222.121.255	Respina Networks & Beyond PJSC
185.222.122.0	185.222.123.255	Poshtkar Rayaneh Kharg Company PJS
185.222.163.0	185.222.163.255	Mizban Dadeh Roham Co
185.222.180.0	185.222.180.255	Iran Zamin Bank PJSC
185.222.181.0	185.222.181.255	Iran Zamin Bank PJSC
185.222.182.0	185.222.182.255	Iran Zamin Bank PJSC
185.222.183.0	185.222.183.255	Iran Zamin Bank PJSC
185.222.210.0	185.222.210.255	Dade Samane Fanava Company (PJS)
185.223.160.0	185.223.160.255	Netafraz Iranian Ltd.
185.224.176.0	185.224.176.255	Rahnamoun Rayaneh Ertebatat Company (Ltd.)
185.224.177.0	185.224.177.255	Rahnamoun Rayaneh Ertebatat Company (Ltd.)
185.224.178.0	185.224.178.255	Rahnamoun Rayaneh Ertebatat Company (Ltd.)
185.224.179.0	185.224.179.255	Rahnamoun Rayaneh Ertebatat Company (Ltd.)
185.225.80.0	185.225.83.255	Hormoz IT & Network Waves Connection Co. (PJS)
185.225.180.0	185.225.183.255	Shahrekord University
185.225.240.0	185.225.240.255	Respina Networks & Beyond PJSC
185.225.241.0	185.225.241.255	Respina Networks & Beyond PJSC
185.225.242.0	185.225.242.255	Respina Networks & Beyond PJSC
185.225.243.0	185.225.243.255	Respina Networks & Beyond PJSC
185.226.97.0	185.226.97.255	Parsway Shomal Company Ltd
185.226.116.0	185.226.119.255	Noyan Abr Arvan Co. ( Private Joint Stock)
185.226.132.0	185.226.135.255	Khalij Fars Ettela Resan LTD
185.226.140.0	185.226.140.255	Saba Abr Mizban LLC
185.226.141.0	185.226.141.255	Saba Abr Mizban LLC
185.226.142.0	185.226.142.255	Saba Abr Mizban LLC
185.226.143.0	185.226.143.255	Saba Abr Mizban LLC
185.227.64.0	185.227.64.255	Saman Kish Electronic Payment Company
185.227.65.0	185.227.65.255	Saman Kish Electronic Payment Company
185.227.66.0	185.227.66.255	Saman Kish Electronic Payment Company
185.227.67.0	185.227.67.255	Saman Kish Electronic Payment Company
185.227.78.0	185.227.78.255	Web Dadeh Paydar Co (Ltd)
185.227.79.0	185.227.79.255	Web Dadeh Paydar Co (Ltd)
185.227.116.0	185.227.116.255	Arya Hamrah Samaneh PJS
185.227.117.0	185.227.117.255	Arya Hamrah Samaneh PJS
185.227.118.0	185.227.119.255	Arya Hamrah Samaneh PJS
185.228.58.0	185.228.58.255	Sindad Network Technology PJSC
185.228.59.0	185.228.59.255	Toesegaran Shabakeh Arseh Novin Ltd
185.228.236.0	185.228.236.255	Noyan Abr Arvan Co. ( Private Joint Stock)
185.228.237.0	185.228.237.255	Noyan Abr Arvan Co. ( Private Joint Stock)
185.228.238.0	185.228.238.255	Noyan Abr Arvan Co. ( Private Joint Stock)
185.228.239.0	185.228.239.255	Noyan Abr Arvan Co. ( Private Joint Stock)
185.229.0.0	185.229.0.255	Bours Kalaye Iran PJSC (Iran Mercantile Exchange)
185.229.1.0	185.229.1.255	Bours Kalaye Iran PJSC (Iran Mercantile Exchange)
185.229.2.0	185.229.3.255	Bours Kalaye Iran PJSC (Iran Mercantile Exchange)
185.229.28.0	185.229.28.255	Atrin Information & Communications Technology Company PJS
185.229.29.0	185.229.29.255	Atrin Information & Communications Technology Company PJS
185.229.30.0	185.229.30.255	Atrin Information & Communications Technology Company PJS
185.229.31.0	185.229.31.255	Atrin Information & Communications Technology Company PJS
185.229.133.0	185.229.133.255	Atrin Information & Communications Technology Company PJS
185.229.134.0	185.229.134.255	Aranik Communications Technology PJSC
185.229.135.0	185.229.135.255	Sindad Network Technology PJSC
185.229.204.0	185.229.204.255	Asre Pardazeshe Ettelaate Amin Institute
185.231.65.0	185.231.65.255	Dadeh Rayanesh Abri Pardis Private Joint Stock Company
185.231.112.0	185.231.112.255	Vandad Vira Hooman LLC
185.231.113.0	185.231.113.255	Patron Technology Persia Ltd
185.231.114.0	185.231.114.255	Amnpardaz Soft Corporation
185.231.115.0	185.231.115.255	Vandad Vira Hooman LLC
185.231.180.0	185.231.183.255	Noyan Abr Arvan Co. ( Private Joint Stock)
185.232.152.0	185.232.152.255	Tebyan-e-Noor Cultural-Artistic Institute
185.232.153.0	185.232.153.255	Tebyan-e-Noor Cultural-Artistic Institute
185.232.154.0	185.232.154.255	Tebyan-e-Noor Cultural-Artistic Institute
185.232.155.0	185.232.155.255	Tebyan-e-Noor Cultural-Artistic Institute
185.232.176.0	185.232.176.255	Export Development Bank of Iran - Public Joint Stock
185.232.177.0	185.232.177.255	Export Development Bank of Iran - Public Joint Stock
185.232.178.0	185.232.178.255	Export Development Bank of Iran - Public Joint Stock
185.232.179.0	185.232.179.255	Export Development Bank of Iran - Public Joint Stock
185.233.12.0	185.233.15.255	Asiatech Data Transmission Company
185.233.84.0	185.233.84.255	Fanavari Hamrah Payda Co. ( Private J.S.)
185.233.85.0	185.233.85.255	Fanavari Hamrah Payda Co. ( Private J.S.)
185.233.86.0	185.233.86.255	Fanavari Hamrah Payda Co. ( Private J.S.)
185.233.87.0	185.233.87.255	Fanavari Hamrah Payda Co. ( Private J.S.)
185.233.131.0	185.233.131.255	Sefroyek Pardaz Engineering PJSC
185.234.14.0	185.234.14.255	Sefroyek Pardaz Engineering PJSC
185.234.192.0	185.234.192.255	Vista Samaneh ASA Company (PJS)
185.234.193.0	185.234.193.255	Vista Samaneh ASA Company (PJS)
185.234.194.0	185.234.194.255	Vista Samaneh ASA Company (PJS)
185.234.195.0	185.234.195.255	Vista Samaneh ASA Company (PJS)
185.235.136.0	185.235.136.255	Cooperative Afra ertebatat-e-sabet-e Rasa Co
185.235.139.0	185.235.139.255	Cooperative Afra ertebatat-e-sabet-e Rasa Co
185.235.196.0	185.235.196.255	Taymaz Gostaran Naghshe Almas Co. ( Ltd )
185.235.197.0	185.235.197.255	Taymaz Gostaran Naghshe Almas Co. ( Ltd )
185.235.198.0	185.235.198.255	Toesegaran Shabakeh Arseh Novin Ltd
185.235.245.0	185.235.245.255	Sefroyek Pardaz Engineering PJSC
185.236.36.0	185.236.36.255	Sindad Network Technology PJSC
185.236.37.0	185.236.37.255	Sindad Network Technology PJSC
185.236.38.0	185.236.38.255	Sindad Network Technology PJSC
185.236.39.0	185.236.39.255	Sindad Network Technology PJSC
185.236.88.0	185.236.88.255	University of Art
185.236.89.0	185.236.89.255	University of Art
185.236.90.0	185.236.91.255	University of Art
185.237.8.0	185.237.11.255	Asiatech Data Transmission Company
185.237.84.0	185.237.84.255	Shabakieh Isfahan Co PJSC
185.237.85.0	185.237.85.255	Shabakieh Isfahan Co PJSC
185.237.86.0	185.237.86.255	Shabakieh Isfahan Co PJSC
185.237.87.0	185.237.87.255	Shabakieh Isfahan Co PJSC
185.238.44.0	185.238.44.255	Hesabgar Pardaz Gharb PJSC
185.238.45.0	185.238.45.255	Hesabgar Pardaz Gharb PJSC
185.238.46.0	185.238.46.255	Hesabgar Pardaz Gharb PJSC
185.238.47.0	185.238.47.255	Hesabgar Pardaz Gharb PJSC
185.238.92.0	185.238.95.255	Rooyekhat Media Company Ltd
185.238.140.0	185.238.140.255	Avaye Segal Net Co. PJS
185.238.143.0	185.238.143.255	Sindad Network Technology PJSC
185.239.0.0	185.239.0.255	Vandad Vira Hooman LLC
185.239.2.0	185.239.2.255	Ferdowsi University of Mashhad
185.239.3.0	185.239.3.255	Vandad Vira Hooman LLC
185.239.104.0	185.239.104.255	Noyan Abr Arvan Co. ( Private Joint Stock)
185.239.105.0	185.239.105.255	Noyan Abr Arvan Co. ( Private Joint Stock)
185.239.106.0	185.239.107.255	Noyan Abr Arvan Co. ( Private Joint Stock)
185.240.56.0	185.240.59.255	Fanavaran Novin Bonyan Pasargad Co. (PSJ)
185.240.148.0	185.240.148.255	Part Financial Information Processing Co.
185.240.149.0	185.240.149.255	Computer Research Center of Islamic Sciences
185.240.150.0	185.240.150.255	Web Gostaran Bandar Company (PJS)
185.240.151.0	185.240.151.255	Part Financial Information Processing Co.
185.241.204.0	185.241.204.255	Web Dadeh Paydar Co (Ltd)
185.243.48.0	185.243.48.255	Vandad Vira Hooman LLC
185.243.49.0	185.243.49.255	Peyman Ertebatat Pouya Company (Ltd)
185.243.50.0	185.243.50.255	Toloe Rayaneh Loghman Educational and Cultural Co. LTD
185.243.51.0	185.243.51.255	Amir Mohammad Masoudi
185.246.4.0	185.246.4.255	Khalij Fars Ettela Resan LTD
185.246.5.0	185.246.5.255	Khalij Fars Ettela Resan LTD
185.246.6.0	185.246.6.255	Khalij Fars Ettela Resan LTD
185.246.7.0	185.246.7.255	Khalij Fars Ettela Resan LTD
185.249.9.0	185.249.9.255	Web Dadeh Paydar Co (Ltd)
185.249.10.0	185.249.10.255	Web Dadeh Paydar Co (Ltd)
185.251.76.0	185.251.76.255	Ertebatdehi Ilam Technical & Services Company PJS
185.251.77.0	185.251.77.255	Ertebatdehi Ilam Technical & Services Company PJS
185.251.78.0	185.251.78.255	Ertebatdehi Ilam Technical & Services Company PJS
185.251.79.0	185.251.79.255	Ertebatdehi Ilam Technical & Services Company PJS
185.252.28.0	185.252.28.255	Shabakeh Ertebatat Artak Towseeh PJSC
185.252.29.0	185.252.29.255	Shabakeh Ertebatat Artak Towseeh PJSC
185.252.30.0	185.252.30.255	Shabakeh Ertebatat Artak Towseeh PJSC
185.252.31.0	185.252.31.255	Shabakeh Ertebatat Artak Towseeh PJSC
185.252.84.0	185.252.84.255	Sindad Network Technology PJSC
185.252.85.0	185.252.85.255	Bazi Gostaran Derick
185.252.86.0	185.252.86.255	Vandad Vira Hooman LLC
185.252.200.0	185.252.200.255	Sefroyek Pardaz Engineering PJSC
185.254.165.0	185.254.165.255	Pishgam Rayan Dadeh Pajooh PJSC
185.254.166.0	185.254.166.255	Kian Iranian Smart Technologies Development Group (PJS)
185.255.88.0	185.255.88.255	Green Web Samaneh Novin PJSC
185.255.89.0	185.255.89.255	Green Web Samaneh Novin PJSC
185.255.90.0	185.255.90.255	Green Web Samaneh Novin PJSC
185.255.91.0	185.255.91.255	Green Web Samaneh Novin PJSC
185.255.98.0	185.255.99.255	Mobin Net Communication Company (Private Joint Stock)
185.255.208.0	185.255.208.255	Cooperative Afra ertebatat-e-sabet-e Rasa Co
185.255.209.0	185.255.209.255	Cooperative Afra ertebatat-e-sabet-e Rasa Co
185.255.210.0	185.255.211.255	Cooperative Afra ertebatat-e-sabet-e Rasa Co
188.0.240.0	188.0.240.255	Asiatech Data Transmission Company
188.0.241.0	188.0.241.255	Asiatech Data Transmission Company
188.0.242.0	188.0.243.255	Iran Telecommunication Company Pjs
188.0.244.0	188.0.244.255	Iran Telecommunication Company Pjs
188.0.245.0	188.0.245.255	Al-Mustafa International University
188.0.246.0	188.0.246.255	Iran Telecommunication Company Pjs
188.0.247.0	188.0.247.255	Iran Telecommunication Company Pjs
188.0.248.0	188.0.255.255	Iran Telecommunication Company Pjs
188.75.64.0	188.75.64.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.75.65.0	188.75.65.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.75.66.0	188.75.67.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.75.68.0	188.75.69.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.75.70.0	188.75.70.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.75.71.0	188.75.71.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.75.72.0	188.75.73.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.75.74.0	188.75.74.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.75.75.0	188.75.75.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.75.76.0	188.75.77.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.75.78.0	188.75.78.255	Rasaneh Avabarid Private Joint Stock Company
188.75.79.0	188.75.79.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.75.80.0	188.75.83.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.75.84.0	188.75.84.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.75.85.0	188.75.85.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.75.86.0	188.75.87.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.75.88.0	188.75.88.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.75.89.0	188.75.89.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.75.90.0	188.75.91.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.75.92.0	188.75.92.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.75.93.0	188.75.93.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.75.94.0	188.75.94.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.75.95.0	188.75.95.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.75.96.0	188.75.97.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.75.98.0	188.75.98.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.75.99.0	188.75.99.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.75.100.0	188.75.101.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.75.102.0	188.75.103.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.75.104.0	188.75.104.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.75.105.0	188.75.105.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.75.106.0	188.75.106.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.75.107.0	188.75.107.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.75.108.0	188.75.108.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.75.109.0	188.75.109.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.75.110.0	188.75.111.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.75.112.0	188.75.112.255	Datis Arian Gheshm Software Company PJSC
188.75.113.0	188.75.113.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.75.114.0	188.75.115.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.75.116.0	188.75.117.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.75.118.0	188.75.119.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.75.120.0	188.75.121.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.75.122.0	188.75.123.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.75.124.0	188.75.125.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.75.126.0	188.75.126.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.75.127.0	188.75.127.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.94.188.0	188.94.188.255	Tose Abzar Bazar Sarmayeh PJSC
188.95.198.0	188.95.198.255	Taliya Communicational Development Company (Private Joint Stock)
188.118.64.0	188.118.67.255	University of Tehran
188.118.68.0	188.118.71.255	University of Tehran
188.118.72.0	188.118.79.255	University of Tehran
188.118.80.0	188.118.81.255	University of Tehran
188.118.82.0	188.118.83.255	University of Tehran
188.118.84.0	188.118.84.255	University of Tehran
188.118.85.0	188.118.85.255	University of Tehran
188.118.86.0	188.118.87.255	University of Tehran
188.118.88.0	188.118.89.255	University of Tehran
188.118.90.0	188.118.90.255	University of Tehran
188.118.91.0	188.118.91.255	University of Tehran
188.118.92.0	188.118.95.255	University of Tehran
188.118.96.0	188.118.127.255	University of Tehran
188.121.96.0	188.121.99.255	Noyan Abr Arvan Co. ( Private Joint Stock)
188.121.100.0	188.121.103.255	Noyan Abr Arvan Co. ( Private Joint Stock)
188.121.104.0	188.121.107.255	Noyan Abr Arvan Co. ( Private Joint Stock)
188.121.108.0	188.121.111.255	Noyan Abr Arvan Co. ( Private Joint Stock)
188.121.112.0	188.121.115.255	Noyan Abr Arvan Co. ( Private Joint Stock)
188.121.116.0	188.121.119.255	Noyan Abr Arvan Co. ( Private Joint Stock)
188.121.120.0	188.121.123.255	Noyan Abr Arvan Co. ( Private Joint Stock)
188.121.124.0	188.121.125.255	Noyan Abr Arvan Co. ( Private Joint Stock)
188.121.126.0	188.121.126.255	Noyan Abr Arvan Co. ( Private Joint Stock)
188.121.127.0	188.121.127.255	Noyan Abr Arvan Co. ( Private Joint Stock)
188.121.128.0	188.121.128.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.121.129.0	188.121.129.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.121.130.0	188.121.131.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.121.132.0	188.121.135.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.121.136.0	188.121.143.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.121.144.0	188.121.147.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.121.148.0	188.121.149.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.121.150.0	188.121.151.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.121.152.0	188.121.155.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.121.156.0	188.121.156.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.121.157.0	188.121.157.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.121.158.0	188.121.159.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.122.96.0	188.122.111.255	Mobile Communication Company of Iran Plc
188.122.112.0	188.122.127.255	Mobile Communication Company of Iran Plc
188.136.128.0	188.136.128.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.136.129.0	188.136.129.255	Ariana Gostar Spadana (PJSC)
188.136.130.0	188.136.130.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.136.131.0	188.136.131.255	Ariana Gostar Spadana (PJSC)
188.136.132.0	188.136.132.255	Ariana Gostar Spadana (PJSC)
188.136.133.0	188.136.133.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.136.134.0	188.136.134.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.136.135.0	188.136.135.255	Ariana Gostar Spadana (PJSC)
188.136.136.0	188.136.136.255	Ariana Gostar Spadana (PJSC)
188.136.137.0	188.136.137.255	Ariana Gostar Spadana (PJSC)
188.136.138.0	188.136.138.255	Ariana Gostar Spadana (PJSC)
188.136.139.0	188.136.139.255	Ariana Gostar Spadana (PJSC)
188.136.140.0	188.136.140.255	Ariana Gostar Spadana (PJSC)
188.136.141.0	188.136.141.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.136.142.0	188.136.142.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.136.143.0	188.136.143.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.136.144.0	188.136.144.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.136.145.0	188.136.145.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.136.146.0	188.136.147.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.136.148.0	188.136.148.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.136.149.0	188.136.149.255	Ariana Gostar Spadana (PJSC)
188.136.150.0	188.136.150.255	Ariana Gostar Spadana (PJSC)
188.136.151.0	188.136.151.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.136.152.0	188.136.152.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.136.153.0	188.136.153.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.136.154.0	188.136.154.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.136.155.0	188.136.155.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.136.156.0	188.136.156.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.136.157.0	188.136.157.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.136.158.0	188.136.158.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.136.159.0	188.136.159.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.136.160.0	188.136.160.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.136.161.0	188.136.161.255	Ariana Gostar Spadana (PJSC)
188.136.162.0	188.136.162.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.136.163.0	188.136.163.255	Ariana Gostar Spadana (PJSC)
188.136.164.0	188.136.164.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.136.165.0	188.136.165.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.136.166.0	188.136.166.255	Ariana Gostar Spadana (PJSC)
188.136.167.0	188.136.167.255	Ariana Gostar Spadana (PJSC)
188.136.168.0	188.136.169.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.136.170.0	188.136.170.255	Ariana Gostar Spadana (PJSC)
188.136.171.0	188.136.171.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.136.172.0	188.136.172.255	Ariana Gostar Spadana (PJSC)
188.136.173.0	188.136.173.255	Ariana Gostar Spadana (PJSC)
188.136.174.0	188.136.174.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.136.175.0	188.136.175.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.136.176.0	188.136.183.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.136.184.0	188.136.184.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.136.185.0	188.136.185.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.136.186.0	188.136.187.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.136.188.0	188.136.189.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.136.190.0	188.136.190.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.136.191.0	188.136.191.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.136.192.0	188.136.193.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.136.194.0	188.136.194.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.136.195.0	188.136.195.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.136.196.0	188.136.196.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.136.197.0	188.136.197.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.136.198.0	188.136.199.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.136.200.0	188.136.200.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.136.201.0	188.136.201.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.136.202.0	188.136.203.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.136.204.0	188.136.207.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.136.208.0	188.136.208.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.136.209.0	188.136.209.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.136.210.0	188.136.211.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.136.212.0	188.136.215.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.136.216.0	188.136.219.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.136.220.0	188.136.221.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.136.222.0	188.136.223.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
188.158.0.0	188.158.255.255	Parvaresh Dadeha Co. Private Joint Stock
188.159.0.0	188.159.63.255	Parvaresh Dadeha Co. Private Joint Stock
188.159.64.0	188.159.95.255	Parvaresh Dadeha Co. Private Joint Stock
188.159.96.0	188.159.127.255	Parvaresh Dadeha Co. Private Joint Stock
188.159.128.0	188.159.159.255	Parvaresh Dadeha Co. Private Joint Stock
188.159.160.0	188.159.191.255	Parvaresh Dadeha Co. Private Joint Stock
188.159.192.0	188.159.223.255	Parvaresh Dadeha Co. Private Joint Stock
188.191.176.0	188.191.183.255	Information Technology Organization of Isfahan Municipality
188.208.56.0	188.208.63.255	Iran Telecommunication Company Pjs
188.208.64.0	188.208.95.255	Rightel Communication Service Company PJS
188.208.144.0	188.208.151.255	Rightel Communication Service Company PJS
188.208.152.0	188.208.155.255	Rightel Communication Service Company PJS
188.208.156.0	188.208.156.255	Rightel Communication Service Company PJS
188.208.157.0	188.208.157.255	Rightel Communication Service Company PJS
188.208.158.0	188.208.159.255	Rightel Communication Service Company PJS
188.208.160.0	188.208.175.255	Rightel Communication Service Company PJS
188.208.176.0	188.208.191.255	Rightel Communication Service Company PJS
188.208.200.0	188.208.201.255	Iran Telecommunication Company Pjs
188.208.202.0	188.208.203.255	Iran Telecommunication Company Pjs
188.208.208.0	188.208.215.255	Iran Telecommunication Company Pjs
188.208.224.0	188.208.255.255	Rightel Communication Service Company PJS
188.209.0.0	188.209.3.255	Iran Telecommunication Company Pjs
188.209.4.0	188.209.7.255	Iran Telecommunication Company Pjs
188.209.8.0	188.209.15.255	Iran Telecommunication Company Pjs
188.209.16.0	188.209.31.255	Iran Telecommunication Company Pjs
188.209.32.0	188.209.47.255	Iran Telecommunication Company Pjs
188.209.64.0	188.209.79.255	Iran Telecommunication Company Pjs
188.209.116.0	188.209.116.255	Fannavaran-e Idea Pardaz-e Saba PJSC
188.209.117.0	188.209.117.255	Fannavaran-e Idea Pardaz-e Saba PJSC
188.209.118.0	188.209.118.255	Fannavaran-e Idea Pardaz-e Saba PJSC
188.209.119.0	188.209.119.255	Fannavaran-e Idea Pardaz-e Saba PJSC
188.209.152.0	188.209.152.255	Sefroyek Pardaz Engineering PJSC
188.209.153.0	188.209.153.255	Sefroyek Pardaz Engineering PJSC
188.209.192.0	188.209.192.255	Mobile Communication Company of Iran Plc
188.209.193.0	188.209.193.255	Mobile Communication Company of Iran Plc
188.209.194.0	188.209.195.255	Mobile Communication Company of Iran Plc
188.209.196.0	188.209.199.255	Mobile Communication Company of Iran Plc
188.209.200.0	188.209.207.255	Mobile Communication Company of Iran Plc
188.210.64.0	188.210.79.255	Mobile Communication Company of Iran Plc
188.210.80.0	188.210.87.255	Iran Telecommunication Company Pjs
188.210.96.0	188.210.127.255	Iran Telecommunication Company Pjs
188.210.128.0	188.210.191.255	Iran Telecommunication Company Pjs
188.210.192.0	188.210.207.255	Mobile Communication Company of Iran Plc
188.210.232.0	188.210.235.255	Iran Telecommunication Company Pjs
188.211.0.0	188.211.15.255	Mobin Net Communication Company (Private Joint Stock)
188.211.32.0	188.211.63.255	Iran Telecommunication Company Pjs
188.211.64.0	188.211.127.255	Iran Telecommunication Company Pjs
188.211.128.0	188.211.159.255	Iran Telecommunication Company Pjs
188.211.176.0	188.211.191.255	Iran Telecommunication Company Pjs
188.211.192.0	188.211.199.255	Iran Telecommunication Company Pjs
188.211.200.0	188.211.207.255	Iran Telecommunication Company Pjs
188.211.208.0	188.211.215.255	Iran Telecommunication Company Pjs
188.211.216.0	188.211.223.255	Iran Telecommunication Company Pjs
188.212.22.0	188.212.22.255	Netmihan Communication Company Ltd
188.212.48.0	188.212.48.255	Mobile Communication Company of Iran Plc
188.212.49.0	188.212.49.255	Mobile Communication Company of Iran Plc
188.212.50.0	188.212.51.255	Mobile Communication Company of Iran Plc
188.212.52.0	188.212.55.255	Mobile Communication Company of Iran Plc
188.212.56.0	188.212.63.255	Mobile Communication Company of Iran Plc
188.212.64.0	188.212.95.255	Iran Telecommunication Company Pjs
188.212.96.0	188.212.96.255	Vandad Vira Hooman LLC
188.212.97.0	188.212.97.255	Ali Monfared
188.212.98.0	188.212.98.255	Padidar Technology PJSC
188.212.99.0	188.212.99.255	FanAvaran Mihan Mizban PJSC
188.212.144.0	188.212.151.255	Iran Telecommunication Company Pjs
188.212.160.0	188.212.191.255	Iran Telecommunication Company Pjs
188.212.200.0	188.212.207.255	Iran Telecommunication Company Pjs
188.212.208.0	188.212.223.255	Iran Telecommunication Company Pjs
188.212.224.0	188.212.239.255	Iran Telecommunication Company Pjs
188.212.240.0	188.212.243.255	Mobin Net Communication Company (Private Joint Stock)
188.212.244.0	188.212.247.255	Mobin Net Communication Company (Private Joint Stock)
188.213.64.0	188.213.64.255	Mobin Net Communication Company (Private Joint Stock)
188.213.65.0	188.213.65.255	Mobin Net Communication Company (Private Joint Stock)
188.213.66.0	188.213.66.255	Mobin Net Communication Company (Private Joint Stock)
188.213.67.0	188.213.67.255	Mobin Net Communication Company (Private Joint Stock)
188.213.68.0	188.213.68.255	Mobin Net Communication Company (Private Joint Stock)
188.213.69.0	188.213.69.255	Mobin Net Communication Company (Private Joint Stock)
188.213.70.0	188.213.70.255	Mobin Net Communication Company (Private Joint Stock)
188.213.71.0	188.213.71.255	Mobin Net Communication Company (Private Joint Stock)
188.213.72.0	188.213.72.255	Mobin Net Communication Company (Private Joint Stock)
188.213.73.0	188.213.73.255	Mobin Net Communication Company (Private Joint Stock)
188.213.74.0	188.213.74.255	Mobin Net Communication Company (Private Joint Stock)
188.213.75.0	188.213.75.255	Mobin Net Communication Company (Private Joint Stock)
188.213.76.0	188.213.79.255	Mobin Net Communication Company (Private Joint Stock)
188.213.96.0	188.213.127.255	Iran Telecommunication Company Pjs
188.213.144.0	188.213.159.255	Iran Telecommunication Company Pjs
188.213.176.0	188.213.191.255	Iran Telecommunication Company Pjs
188.213.192.0	188.213.195.255	Mobin Net Communication Company (Private Joint Stock)
188.213.196.0	188.213.199.255	Noyan Abr Arvan Co. ( Private Joint Stock)
188.213.208.0	188.213.209.255	Iran Telecommunication Company Pjs
188.213.210.0	188.213.211.255	Iran Telecommunication Company Pjs
188.214.4.0	188.214.5.255	Rightel Communication Service Company PJS
188.214.6.0	188.214.6.255	Rightel Communication Service Company PJS
188.214.7.0	188.214.7.255	Rightel Communication Service Company PJS
188.214.84.0	188.214.87.255	Rightel Communication Service Company PJS
188.214.96.0	188.214.99.255	Iran Telecommunication Company Pjs
188.214.120.0	188.214.121.255	Parvaresh Dadeha Co. Private Joint Stock
188.214.160.0	188.214.191.255	Iran Telecommunication Company Pjs
188.214.216.0	188.214.223.255	Iran Telecommunication Company Pjs
188.215.24.0	188.215.27.255	Rightel Communication Service Company PJS
188.215.88.0	188.215.91.255	Rightel Communication Service Company PJS
188.215.128.0	188.215.143.255	Iran Telecommunication Company Pjs
188.215.160.0	188.215.175.255	Iran Telecommunication Company Pjs
188.215.176.0	188.215.191.255	Iran Telecommunication Company Pjs
188.215.192.0	188.215.223.255	Iran Telecommunication Company Pjs
188.215.240.0	188.215.241.255	Iran Telecommunication Company Pjs
188.215.242.0	188.215.243.255	Iran Telecommunication Company Pjs
188.229.0.0	188.229.15.255	Mobile Communication Company of Iran Plc
188.229.16.0	188.229.31.255	Mobile Communication Company of Iran Plc
188.229.32.0	188.229.47.255	Mobile Communication Company of Iran Plc
188.229.48.0	188.229.63.255	Mobile Communication Company of Iran Plc
188.229.64.0	188.229.79.255	Mobile Communication Company of Iran Plc
188.229.80.0	188.229.95.255	Mobile Communication Company of Iran Plc
188.229.96.0	188.229.111.255	Mobile Communication Company of Iran Plc
188.229.112.0	188.229.115.255	Mobile Communication Company of Iran Plc
188.229.116.0	188.229.116.255	Mobile Communication Company of Iran Plc
188.229.117.0	188.229.117.255	Mobile Communication Company of Iran Plc
188.229.118.0	188.229.119.255	Mobile Communication Company of Iran Plc
188.229.120.0	188.229.127.255	Mobile Communication Company of Iran Plc
188.240.196.0	188.240.196.255	Vandad Vira Hooman LLC
188.240.212.0	188.240.212.255	Padidar Technology PJSC
188.240.248.0	188.240.255.255	Iran Telecommunication Company Pjs
188.253.2.0	188.253.2.255	Avini Cultural and Art Institute
188.253.3.0	188.253.3.255	Avini Cultural and Art Institute
188.253.32.0	188.253.63.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
188.253.64.0	188.253.95.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
192.15.0.0	192.15.0.255	Mobile Communication Company of Iran Plc
192.15.1.0	192.15.1.255	Mobile Communication Company of Iran Plc
192.15.2.0	192.15.3.255	Mobile Communication Company of Iran Plc
192.15.4.0	192.15.7.255	Mobile Communication Company of Iran Plc
192.15.8.0	192.15.15.255	Mobile Communication Company of Iran Plc
192.15.16.0	192.15.31.255	Mobile Communication Company of Iran Plc
192.15.32.0	192.15.47.255	Mobile Communication Company of Iran Plc
192.15.48.0	192.15.63.255	Mobile Communication Company of Iran Plc
192.15.64.0	192.15.71.255	Mobile Communication Company of Iran Plc
192.15.72.0	192.15.75.255	Mobile Communication Company of Iran Plc
192.15.76.0	192.15.77.255	Mobile Communication Company of Iran Plc
192.15.78.0	192.15.78.255	Mobile Communication Company of Iran Plc
192.15.79.0	192.15.79.255	Mobile Communication Company of Iran Plc
192.15.80.0	192.15.95.255	Mobile Communication Company of Iran Plc
192.15.96.0	192.15.111.255	Mobile Communication Company of Iran Plc
192.15.112.0	192.15.127.255	Mobile Communication Company of Iran Plc
192.15.128.0	192.15.128.255	Mobile Communication Company of Iran Plc
192.15.129.0	192.15.129.255	Mobile Communication Company of Iran Plc
192.15.130.0	192.15.131.255	Mobile Communication Company of Iran Plc
192.15.132.0	192.15.135.255	Mobile Communication Company of Iran Plc
192.15.136.0	192.15.143.255	Mobile Communication Company of Iran Plc
192.15.144.0	192.15.159.255	Mobile Communication Company of Iran Plc
192.15.160.0	192.15.175.255	Mobile Communication Company of Iran Plc
192.15.176.0	192.15.191.255	Mobile Communication Company of Iran Plc
192.15.192.0	192.15.207.255	Mobile Communication Company of Iran Plc
192.15.208.0	192.15.223.255	Mobile Communication Company of Iran Plc
192.15.224.0	192.15.239.255	Mobile Communication Company of Iran Plc
192.15.240.0	192.15.255.255	Mobile Communication Company of Iran Plc
192.166.36.0	192.166.36.255	Hessam Isaei Trading As Mahallat Network
192.166.37.0	192.166.37.255	Sabz Pardazesh System Ghadir Co. (P.J.S.)
192.166.38.0	192.166.38.255	Amin Asia Cloud Data PJSC
193.0.156.0	193.0.156.255	Iran Grid Management Company (IGMC)
193.3.31.0	193.3.31.255	Barmak Khotvaneh
193.3.182.0	193.3.182.255	Saba Abr Mizban LLC
193.3.231.0	193.3.231.255	Noavaran Shabakeh Sabz Mehregan (Ltd.)
193.3.255.0	193.3.255.255	Towse'eh Sarmayeh Gozari Entekhab Group PJSC
193.5.44.0	193.5.44.255	Vandad Vira Hooman LLC
193.8.139.0	193.8.139.255	Saman Bank PJSC
193.9.24.0	193.9.24.255	Fanavaran Mehr Vatan Tehran Server Group LLC
193.19.144.0	193.19.144.255	Ertebatat Sabet Parsian Co. PJS
193.19.145.0	193.19.145.255	Ertebatat Sabet Parsian Co. PJS
193.19.147.0	193.19.147.255	Peimayesh Ertebat Oxin Co.(PJSC)
193.22.20.0	193.22.20.255	Shahid Chamran University of Ahvaz
193.24.105.0	193.24.105.255	Web Dadeh Paydar Co (Ltd)
193.24.118.0	193.24.118.255	Sepehr Saye Ban Shokoufaei Company (LTD)
193.24.120.0	193.24.120.255	Web Dadeh Paydar Co (Ltd)
193.27.9.0	193.27.9.255	Respina Networks & Beyond PJSC
193.28.181.0	193.28.181.255	Public JSC Bank Sepah
193.29.26.0	193.29.26.255	Hezardastan Unit Cloud Computing PJSC
193.30.30.0	193.30.30.255	Mobin Net Communication Company (Private Joint Stock)
193.32.80.0	193.32.81.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
193.34.244.0	193.34.244.255	Tadbir Pardaz IT Group LTD
193.34.245.0	193.34.245.255	Tadbir Pardaz IT Group LTD
193.34.246.0	193.34.246.255	Tadbir Pardaz IT Group LTD
193.34.247.0	193.34.247.255	Tadbir Pardaz IT Group LTD
193.35.62.0	193.35.62.255	Saman Insurance Co. (Public Joint Stock)
193.35.230.0	193.35.230.255	Fannavari Etelaate Samane Kavoshgar Ide Ltd.
193.36.92.0	193.36.92.255	Web Dadeh Paydar Co (Ltd)
193.36.93.0	193.36.93.255	Web Dadeh Paydar Co (Ltd)
193.37.37.0	193.37.37.255	Saba Abr Mizban LLC
193.37.38.0	193.37.38.255	Saba Hour Yeganeh Co. ( Private Joint Stock)
193.38.247.0	193.38.247.255	Behsazan Mellat Company PJS
193.39.70.0	193.39.70.255	Behvar Engineering Design Co.
193.46.214.0	193.46.214.255	Mobin Net Communication Company (Private Joint Stock)
193.56.59.0	193.56.59.255	Farzanegan Pars Communications Company PJS
193.56.61.0	193.56.61.255	Farzanegan Pars Communications Company PJS
193.56.107.0	193.56.107.255	Farzanegan Pars Communications Company PJS
193.56.118.0	193.56.118.255	Farzanegan Pars Communications Company PJS
193.56.181.0	193.56.181.255	Saman Bank PJSC
193.84.255.0	193.84.255.255	Taymaz Gostaran Naghshe Almas Co. ( Ltd )
193.93.169.0	193.93.169.255	Dadeh Gostar Sina Shiraz PJSC
193.93.171.0	193.93.171.255	Pishgam Rayan Dadeh Pajooh PJSC
193.104.22.0	193.104.22.255	Rahyab Rayaneh Gostar
193.104.29.0	193.104.29.255	Fara Ertebat Mabna PJS
193.104.212.0	193.104.212.255	Institute for Research in Fundamental Sciences
193.105.2.0	193.105.2.255	Engeneering Designing and Supply the parts of Iran Khodro PJSC
193.105.6.0	193.105.6.255	Toesegaran Shabakeh Arseh Novin Ltd
193.105.234.0	193.105.234.255	Vandad Vira Hooman LLC
193.106.190.0	193.106.190.255	Abramad Technological Infrastructures Development Company PJS
193.107.48.0	193.107.48.255	Chakavak Pardazesh Hooshmand Alborz
193.109.56.0	193.109.56.255	Ali Norouzi
193.111.234.0	193.111.234.255	Asiatech Data Transmission Company
193.111.235.0	193.111.235.255	Enteghal Dadeh Araz Shomal Company PJSC
193.111.236.0	193.111.236.255	Ettelaat Fanavarn-E Tookan Co Ltd
193.134.100.0	193.134.100.255	Asiatech Data Transmission Company
193.134.101.0	193.134.101.255	Enteghal Dadeh Araz Shomal Company PJSC
193.141.64.0	193.141.64.255	Green Web Samaneh Novin PJSC
193.141.65.0	193.141.65.255	Green Web Samaneh Novin PJSC
193.141.126.0	193.141.126.255	Green Web Samaneh Novin PJSC
193.141.127.0	193.141.127.255	Green Web Samaneh Novin PJSC
193.142.232.0	193.142.233.255	Netafraz Iranian Ltd.
193.142.254.0	193.142.255.255	Netafraz Iranian Ltd.
193.148.64.0	193.148.67.255	Ertebatat Fara Gostar Shargh PJSC
193.150.66.0	193.150.66.255	Parvaz System Information Technology Company (LTD)
193.151.128.0	193.151.131.255	Asre Dadeha Asiatech
193.151.132.0	193.151.135.255	Asre Dadeha Asiatech
193.151.136.0	193.151.139.255	Asre Dadeha Asiatech
193.151.140.0	193.151.143.255	Asre Dadeha Asiatech
193.151.144.0	193.151.147.255	Asre Dadeha Asiatech
193.151.148.0	193.151.151.255	Asre Dadeha Asiatech
193.151.152.0	193.151.155.255	Asre Dadeha Asiatech
193.151.156.0	193.151.156.255	Asre Dadeha Asiatech
193.151.157.0	193.151.157.255	Asre Dadeha Asiatech
193.151.158.0	193.151.158.255	Asre Dadeha Asiatech
193.151.159.0	193.151.159.255	Asre Dadeha Asiatech
193.162.129.0	193.162.129.255	Noavaran Shabakeh Sabz Mehregan (Ltd.)
193.176.97.0	193.176.97.255	Gostaresh Ertebat Azin Kia Company PJSC
193.176.240.0	193.176.241.255	Noyan Abr Arvan Co. ( Private Joint Stock)
193.176.242.0	193.176.243.255	Noyan Abr Arvan Co. ( Private Joint Stock)
193.177.242.0	193.177.242.255	Enteghal Dadeh Araz Shomal Company PJSC
193.177.245.0	193.177.245.255	Reza Salehi
193.178.200.0	193.178.200.255	Asmanfaraz Sepahan Company (P.J.S)
193.178.201.0	193.178.201.255	Lesun Communication Furtherance Engineers Co (Ltd.)
193.178.202.0	193.178.202.255	Lesun Communication Furtherance Engineers Co (Ltd.)
193.178.203.0	193.178.203.255	Asmanfaraz Sepahan Company (P.J.S)
193.186.32.0	193.186.32.255	Parsian High Tech Company PJSC
193.186.215.0	193.186.215.255	Kimiyagaran Padideh Modern Company LLC
193.189.122.0	193.189.122.255	Institute for Research in Fundamental Sciences
193.189.123.0	193.189.123.255	Institute for Research in Fundamental Sciences
193.200.148.0	193.200.148.255	Pars Data Processing Company (LTD)
193.201.23.0	193.201.23.255	Shakiba Teb Medical Technology LLC
193.201.66.0	193.201.66.255	Web Dadeh Paydar Co (Ltd)
193.201.192.0	193.201.192.255	Behpardaz Jahan PJSC
193.201.193.0	193.201.193.255	Behpardaz Jahan PJSC
193.201.194.0	193.201.194.255	Behpardaz Jahan PJSC
193.201.195.0	193.201.195.255	Behpardaz Jahan PJSC
193.222.51.0	193.222.51.255	Saman Insurance Co. (Public Joint Stock)
193.228.90.0	193.228.90.255	Mizban Dade Shetaban Co. (Ltd)
193.228.91.0	193.228.91.255	Mizban Dadeh Pardis .Ltd
193.228.136.0	193.228.136.255	Rayankadeh Apadana Company Ltd
193.228.168.0	193.228.168.255	Mizban Dadeh Pardis .Ltd
193.228.169.0	193.228.169.255	Mizban Dadeh Pardis .Ltd
193.242.125.0	193.242.125.255	Vandad Vira Hooman LLC
193.242.194.0	193.242.195.255	Rayaneh Asr-e Ertebat Bam Technology and Engineering Cooperative Co
193.242.208.0	193.242.208.255	Pars Shabakeh Azarakhsh LLC
193.246.174.0	193.246.174.255	Hesabgar Pardaz Gharb PJSC
193.246.175.0	193.246.175.255	Hesabgar Pardaz Gharb PJSC
193.246.200.0	193.246.200.255	Hesabgar Pardaz Gharb PJSC
193.246.201.0	193.246.201.255	Hesabgar Pardaz Gharb PJSC
194.1.155.0	194.1.155.255	Taymaz Gostaran Naghshe Almas Co. ( Ltd )
194.5.40.0	194.5.43.255	Razavi Hospital
194.5.50.0	194.5.50.255	Web Dadeh Paydar Co (Ltd)
194.5.175.0	194.5.175.255	Vandad Vira Hooman LLC
194.5.176.0	194.5.176.255	Javid Berbid Mamasani Information Technology Company PJS
194.5.177.0	194.5.177.255	Javid Berbid Mamasani Information Technology Company PJS
194.5.178.0	194.5.178.255	Javid Berbid Mamasani Information Technology Company PJS
194.5.179.0	194.5.179.255	Javid Berbid Mamasani Information Technology Company PJS
194.5.188.0	194.5.188.255	Vandad Vira Hooman LLC
194.5.195.0	194.5.195.255	Vandad Vira Hooman LLC
194.5.205.0	194.5.205.255	Vandad Vira Hooman LLC
194.9.56.0	194.9.57.255	Gostaresh Ertebat Azin Kia Company PJSC
194.9.80.0	194.9.81.255	Gostaresh Ertebat Azin Kia Company PJSC
194.26.64.0	194.26.64.255	Padiz Dadeh Resan Pjsc
194.26.99.0	194.26.99.255	Parsway Shomal Company Ltd
194.26.117.0	194.26.117.255	Mehr Ava Gostar Parsian Information Engineering Co. Ltd
194.31.108.0	194.31.108.255	Avaye Segal Net Co. PJS
194.32.209.0	194.32.209.255	Namavaran E Asr E Novin Information Technology (Fanhab) Private Joint-Stock Co
194.32.213.0	194.32.213.255	Web Dadeh Paydar Co (Ltd)
194.32.215.0	194.32.215.255	Web Dadeh Paydar Co (Ltd)
194.33.105.0	194.33.105.255	Pars Shabakeh Azarakhsh LLC
194.33.106.0	194.33.106.255	Ertebatat Sabet Parsian Co. PJS
194.33.107.0	194.33.107.255	Tose'Eh Ertebatat Novin Aria Co PJS
194.33.122.0	194.33.123.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
194.33.124.0	194.33.124.255	Sky Dragon Company With Limited Liability
194.33.125.0	194.33.125.255	Netafraz Iranian Ltd.
194.33.126.0	194.33.127.255	Netafraz Iranian Ltd.
194.34.160.0	194.34.160.255	Hezardastan Unit Cloud Computing PJSC
194.34.161.0	194.34.161.255	Hezardastan Unit Cloud Computing PJSC
194.34.163.0	194.34.163.255	Hezardastan Unit Cloud Computing PJSC
194.36.0.0	194.36.0.255	Day Telecom Company (Ltd.)
194.36.174.0	194.36.175.255	Patron Technology Persia Ltd
194.39.36.0	194.39.36.255	Middle East Bank (PJSC)
194.39.37.0	194.39.37.255	Middle East Bank (PJSC)
194.39.38.0	194.39.38.255	Middle East Bank (PJSC)
194.39.39.0	194.39.39.255	Middle East Bank (PJSC)
194.39.248.0	194.39.248.255	Web Dadeh Paydar Co (Ltd)
194.39.254.0	194.39.254.255	Parsian Technology Innovative Solution Co. PJS.
194.41.48.0	194.41.48.255	Mizban Abri Iman Iranian LLC
194.41.49.0	194.41.49.255	Mizban Abri Iman Iranian LLC
194.41.50.0	194.41.50.255	Mizban Abri Iman Iranian LLC
194.41.51.0	194.41.51.255	Mizban Abri Iman Iranian LLC
194.48.198.0	194.48.198.255	Vandad Vira Hooman LLC
194.50.42.0	194.50.42.255	Tose'eh Lian Yas Ertebat PJSC
194.50.204.0	194.50.204.255	Sourena Pardazesh Arya Software Co. LLC
194.50.209.0	194.50.209.255	Sourena Pardazesh Arya Software Co. LLC
194.50.216.0	194.50.216.255	Sourena Pardazesh Arya Software Co. LLC
194.50.218.0	194.50.218.255	Sourena Pardazesh Arya Software Co. LLC
194.53.118.0	194.53.118.255	Fereidoon Hedayatinia
194.53.119.0	194.53.119.255	Fereidoon Hedayatinia
194.53.122.0	194.53.123.255	Fereidoon Hedayatinia
194.56.148.0	194.56.148.255	Sefroyek Pardaz Engineering PJSC
194.59.170.0	194.59.170.255	Mizban Dade Pasargad LLC
194.59.171.0	194.59.171.255	Mizban Dade Pasargad LLC
194.59.214.0	194.59.214.255	Toosee Ertebatat Damavand
194.59.215.0	194.59.215.255	Ettelaat Fanavarn-E Tookan Co Ltd
194.60.208.0	194.60.208.255	Farzanegan Pars Communications Company PJS
194.60.209.0	194.60.209.255	Farzanegan Pars Communications Company PJS
194.60.210.0	194.60.210.255	Farzanegan Pars Communications Company PJS
194.60.211.0	194.60.211.255	Farzanegan Pars Communications Company PJS
194.60.230.0	194.60.230.255	Pars Shabakeh Azarakhsh LLC
194.60.231.0	194.60.231.255	Pars Shabakeh Azarakhsh LLC
194.62.17.0	194.62.17.255	Faraso Samaneh Pasargad Ltd.
194.62.43.0	194.62.43.255	Vandad Vira Hooman LLC
194.107.116.0	194.107.116.255	Abr Baran Information Technologists LLC
194.110.24.0	194.110.24.255	Web Dadeh Paydar Co (Ltd)
194.110.118.0	194.110.118.255	Gostaresh Ertebat Azin Kia Company PJSC
194.117.64.0	194.117.64.255	Vandad Vira Hooman LLC
194.117.82.0	194.117.82.255	Milad Ghorbani Chahardeh
194.143.140.0	194.143.140.255	National Iranian Oil Company PJSC
194.143.141.0	194.143.141.255	National Iranian Oil Company PJSC
194.145.119.0	194.145.119.255	Respina Networks & Beyond PJSC
194.146.68.0	194.146.68.255	Web Dadeh Paydar Co (Ltd)
194.146.69.0	194.146.69.255	Web Dadeh Paydar Co (Ltd)
194.146.148.0	194.146.148.255	Isfahan University of Technology
194.146.149.0	194.146.149.255	Isfahan University of Technology
194.146.150.0	194.146.150.255	Isfahan University of Technology
194.146.151.0	194.146.151.255	Isfahan University of Technology
194.146.236.0	194.146.236.255	Day Telecom Company (Ltd.)
194.146.237.0	194.146.237.255	Day Telecom Company (Ltd.)
194.146.238.0	194.146.238.255	Day Telecom Company (Ltd.)
194.146.239.0	194.146.239.255	Day Telecom Company (Ltd.)
194.147.142.0	194.147.142.255	Vandad Vira Hooman LLC
194.147.164.0	194.147.164.255	Cooperative Afra ertebatat-e-sabet-e Rasa Co
194.147.165.0	194.147.165.255	Cooperative Afra ertebatat-e-sabet-e Rasa Co
194.147.166.0	194.147.166.255	Cooperative Afra ertebatat-e-sabet-e Rasa Co
194.147.167.0	194.147.167.255	Cooperative Afra ertebatat-e-sabet-e Rasa Co
194.147.170.0	194.147.170.255	Patron Technology Persia Ltd
194.147.212.0	194.147.212.255	Sindad Network Technology PJSC
194.147.222.0	194.147.222.255	Sindad Network Technology PJSC
194.150.68.0	194.150.69.255	Gostaresh Ertebat Azin Kia Company PJSC
194.150.70.0	194.150.71.255	Gostaresh Ertebat Azin Kia Company PJSC
194.150.165.0	194.150.165.255	Hamed Bagheri
194.156.77.0	194.156.77.255	Pishro Fannavaran-E Hooshmand Ravesh JSC
194.156.140.0	194.156.140.255	Golrang System Co. P.J.S
194.156.141.0	194.156.141.255	Golrang System Co. P.J.S
194.156.142.0	194.156.142.255	Golrang System Co. P.J.S
194.156.143.0	194.156.143.255	Golrang System Co. P.J.S
194.180.11.0	194.180.11.255	Vandad Vira Hooman LLC
194.180.209.0	194.180.209.255	Patron Technology Persia Ltd
194.180.224.0	194.180.224.255	Patron Technology Persia Ltd
194.180.238.0	194.180.238.255	Mobin Net Communication Company (Private Joint Stock)
194.225.0.0	194.225.15.255	Institute for Research in Fundamental Sciences
194.225.16.0	194.225.23.255	Institute for Research in Fundamental Sciences
194.225.24.0	194.225.31.255	Shahid Beheshti University
194.225.32.0	194.225.47.255	Institute for Research in Fundamental Sciences
194.225.48.0	194.225.49.255	Tehran University of Medical Science
194.225.50.0	194.225.50.255	Tehran University of Medical Science
194.225.51.0	194.225.51.255	Tehran University of Medical Science
194.225.52.0	194.225.52.255	Tehran University of Medical Science
194.225.53.0	194.225.53.255	Tehran University of Medical Science
194.225.54.0	194.225.55.255	Tehran University of Medical Science
194.225.56.0	194.225.57.255	Tehran University of Medical Science
194.225.58.0	194.225.58.255	Tehran University of Medical Science
194.225.59.0	194.225.59.255	Tehran University of Medical Science
194.225.60.0	194.225.63.255	Tehran University of Medical Science
194.225.64.0	194.225.67.255	Institute for Research in Fundamental Sciences
194.225.68.0	194.225.69.255	Institute for Research in Fundamental Sciences
194.225.70.0	194.225.70.255	Institute for Research in Fundamental Sciences
194.225.71.0	194.225.71.255	Institute for Research in Fundamental Sciences
194.225.72.0	194.225.75.255	Institute for Research in Fundamental Sciences
194.225.76.0	194.225.76.255	Institute for Research in Fundamental Sciences
194.225.77.0	194.225.77.255	Ministry of Health Treatment and Medical Education
194.225.78.0	194.225.78.255	Tehran University of Medical Science
194.225.79.0	194.225.79.255	Tehran University of Medical Science
194.225.80.0	194.225.95.255	Institute for Research in Fundamental Sciences
194.225.96.0	194.225.99.255	Institute for Research in Fundamental Sciences
194.225.100.0	194.225.100.255	Institute for Research in Fundamental Sciences
194.225.101.0	194.225.101.255	IsIran
194.225.102.0	194.225.103.255	Institute for Research in Fundamental Sciences
194.225.104.0	194.225.111.255	Institute for Research in Fundamental Sciences
194.225.112.0	194.225.127.255	Institute for Research in Fundamental Sciences
194.225.128.0	194.225.129.255	Institute for Research in Fundamental Sciences
194.225.130.0	194.225.131.255	Institute for Research in Fundamental Sciences
194.225.132.0	194.225.135.255	Institute for Research in Fundamental Sciences
194.225.136.0	194.225.143.255	Institute for Research in Fundamental Sciences
194.225.144.0	194.225.147.255	Institute for Research in Fundamental Sciences
194.225.148.0	194.225.148.255	Iranian Presidential Administration
194.225.149.0	194.225.149.255	Iranian Presidential Administration
194.225.150.0	194.225.150.255	Institute for Research in Fundamental Sciences
194.225.151.0	194.225.151.255	Institute for Research in Fundamental Sciences
194.225.152.0	194.225.159.255	Institute for Research in Fundamental Sciences
194.225.160.0	194.225.163.255	Institute for Research in Fundamental Sciences
194.225.164.0	194.225.165.255	Institute for Research in Fundamental Sciences
194.225.166.0	194.225.166.255	Tarbiat Modares University
194.225.167.0	194.225.167.255	Tarbiat Modares University
194.225.168.0	194.225.168.255	Tarbiat Modares University
194.225.169.0	194.225.169.255	Tarbiat Modares University
194.225.170.0	194.225.170.255	Tarbiat Modares University
194.225.171.0	194.225.171.255	Tarbiat Modares University
194.225.172.0	194.225.172.255	Tarbiat Modares University
194.225.173.0	194.225.173.255	Tarbiat Modares University
194.225.174.0	194.225.175.255	Institute for Research in Fundamental Sciences
194.225.176.0	194.225.183.255	Institute for Research in Fundamental Sciences
194.225.184.0	194.225.191.255	Iran University of Medical Sciences
194.225.192.0	194.225.199.255	Tehran University of Medical Science
194.225.200.0	194.225.203.255	Tehran University of Medical Science
194.225.204.0	194.225.205.255	Tehran University of Medical Science
194.225.206.0	194.225.206.255	Tehran University of Medical Science
194.225.207.0	194.225.207.255	Tehran University of Medical Science
194.225.208.0	194.225.211.255	Tehran University of Medical Science
194.225.212.0	194.225.213.255	Tehran University of Medical Science
194.225.214.0	194.225.214.255	Tehran University of Medical Science
194.225.215.0	194.225.215.255	Tehran University of Medical Science
194.225.216.0	194.225.223.255	Tehran University of Medical Science
194.225.224.0	194.225.227.255	Iran University of Science and Technology
194.225.228.0	194.225.229.255	Iran University of Science and Technology
194.225.230.0	194.225.230.255	Iran University of Science and Technology
194.225.231.0	194.225.231.255	Iran University of Science and Technology
194.225.232.0	194.225.232.255	Iran University of Science and Technology
194.225.233.0	194.225.233.255	Iran University of Science and Technology
194.225.234.0	194.225.235.255	Iran University of Science and Technology
194.225.236.0	194.225.239.255	Iran University of Science and Technology
194.225.240.0	194.225.255.255	Institute for Research in Fundamental Sciences
194.242.22.0	194.242.22.255	Mobin Net Communication Company (Private Joint Stock)
195.2.234.0	195.2.234.255	Araax Dadeh Gostar Information and Communication Development Co (Private Joint Stock)
195.8.102.0	195.8.102.255	Taknet Afzar Aftab PJSC
195.8.110.0	195.8.110.255	Hesabgar Pardaz Gharb PJSC
195.8.112.0	195.8.112.255	Hesabgar Pardaz Gharb PJSC
195.8.114.0	195.8.114.255	Hesabgar Pardaz Gharb PJSC
195.10.220.0	195.10.220.255	Web Dadeh Paydar Co (Ltd)
195.18.10.0	195.18.10.255	Patron Technology Persia Ltd
195.20.136.0	195.20.136.255	Tehran Regional Electricity Joint Stock Company
195.24.233.0	195.24.233.255	Respina Networks & Beyond PJSC
195.26.27.0	195.26.27.255	Web Dadeh Paydar Co (Ltd)
195.28.10.0	195.28.10.255	Netafraz Iranian Ltd.
195.28.11.0	195.28.11.255	Web Gostaran Bandar Company (PJS)
195.28.168.0	195.28.168.255	Netafraz Iranian Ltd.
195.28.169.0	195.28.169.255	Netafraz Iranian Ltd.
195.62.4.0	195.62.4.255	Andisheh Sabz Mahan Co. Ltd
195.88.188.0	195.88.188.255	Part Mehr Iranian High-tech and Telecommunication Cooperative
195.88.189.0	195.88.189.255	Gostaresh Ertebat Azin Kia Company PJSC
195.88.208.0	195.88.208.255	Vandad Vira Hooman LLC
195.96.128.0	195.96.128.255	Sabz Pardazesh System Ghadir Co. (P.J.S.)
195.96.135.0	195.96.135.255	Giti Secure Cloud LLC
195.96.153.0	195.96.153.255	Karno Tejarat Yasin Company PJS
195.110.38.0	195.110.39.255	Sefroyek Pardaz Engineering PJSC
195.114.4.0	195.114.4.255	Sindad Network Technology PJSC
195.114.5.0	195.114.5.255	Sindad Network Technology PJSC
195.114.8.0	195.114.8.255	Sindad Network Technology PJSC
195.114.9.0	195.114.9.255	Sindad Network Technology PJSC
195.137.167.0	195.137.167.255	Beh Pardakht Melat Co PJS
195.146.32.0	195.146.32.255	Iran Information Technology Company PJSC
195.146.33.0	195.146.33.255	Iran Information Technology Company PJSC
195.146.34.0	195.146.35.255	Iran Information Technology Company PJSC
195.146.36.0	195.146.39.255	Iran Information Technology Company PJSC
195.146.40.0	195.146.47.255	Iran Information Technology Company PJSC
195.146.48.0	195.146.49.255	Iran Information Technology Company PJSC
195.146.50.0	195.146.50.255	Iran Telecommunication Company Pjs
195.146.51.0	195.146.51.255	Iran Information Technology Company PJSC
195.146.52.0	195.146.55.255	Iran Information Technology Company PJSC
195.146.56.0	195.146.57.255	Iran Information Technology Company PJSC
195.146.58.0	195.146.58.255	Iran Information Technology Company PJSC
195.146.59.0	195.146.59.255	Iran Information Technology Company PJSC
195.146.60.0	195.146.60.255	Iran Information Technology Company PJSC
195.146.61.0	195.146.61.255	Iran Information Technology Company PJSC
195.146.62.0	195.146.62.255	Iran Information Technology Company PJSC
195.146.63.0	195.146.63.255	Telecommunication Infrastructure Company
195.149.127.0	195.149.127.255	Mobin Net Communication Company (Private Joint Stock)
195.158.230.0	195.158.230.255	Web Dadeh Paydar Co (Ltd)
195.177.255.0	195.177.255.255	Vandad Vira Hooman LLC
195.181.0.0	195.181.31.255	Iran Telecommunication Company Pjs
195.181.32.0	195.181.47.255	Iran Telecommunication Company Pjs
195.181.48.0	195.181.55.255	Iran Telecommunication Company Pjs
195.181.56.0	195.181.63.255	Iran Telecommunication Company Pjs
195.181.64.0	195.181.79.255	Iran Telecommunication Company Pjs
195.181.80.0	195.181.95.255	Iran Telecommunication Company Pjs
195.181.96.0	195.181.111.255	Iran Telecommunication Company Pjs
195.181.112.0	195.181.119.255	Iran Telecommunication Company Pjs
195.181.120.0	195.181.127.255	Iran Telecommunication Company Pjs
195.182.38.0	195.182.38.255	Padideh Sazan Poshtvar Co. PJS
195.190.139.0	195.190.139.255	Abr Afzar Sedna Ertebat Company (LTD.)
195.190.144.0	195.190.144.255	Pars Abr Toseeh Ertebatat Ltd
195.191.22.0	195.191.22.255	Hesabgar Pardaz Gharb PJSC
195.191.23.0	195.191.23.255	Hesabgar Pardaz Gharb PJSC
195.191.44.0	195.191.44.255	Hesabgar Pardaz Gharb PJSC
195.191.45.0	195.191.45.255	Hesabgar Pardaz Gharb PJSC
195.191.74.0	195.191.74.255	Iranian Blood Transfusion Organization (IBTO)
195.191.75.0	195.191.75.255	Iranian Blood Transfusion Organization (IBTO)
195.200.76.0	195.200.76.255	Saeed Boroumandfar
195.200.77.0	195.200.77.255	Navid Bakzadegan Tabrizi
195.211.44.0	195.211.44.255	Sefroyek Pardaz Engineering PJSC
195.211.45.0	195.211.45.255	Sefroyek Pardaz Engineering PJSC
195.211.46.0	195.211.46.255	Sefroyek Pardaz Engineering PJSC
195.211.47.0	195.211.47.255	Sefroyek Pardaz Engineering PJSC
195.211.71.0	195.211.71.255	Mizban Pardazesh Pouyan LLC
195.214.235.0	195.214.235.255	Zarin Amol Gozar Technology Development Co. Ltd
195.225.232.0	195.225.232.255	Towse'e Saman Information Technology Co. PJS
195.226.223.0	195.226.223.255	Faraso Samaneh Pasargad Ltd.
195.230.97.0	195.230.97.255	Hesabgar Pardaz Gharb PJSC
195.230.105.0	195.230.105.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
195.230.107.0	195.230.107.255	Taknet Afzar Aftab PJSC
195.230.124.0	195.230.124.255	Pishgaman Toseeh Ertebatat Company (Private Joint Stock)
195.234.80.0	195.234.80.255	Web Gostaran Shaina Shiraz LLC
195.234.191.0	195.234.191.255	Maya Pardazesh Iranian LLC
195.238.231.0	195.238.231.255	Web Gostaran Bandar Company (PJS)
195.238.240.0	195.238.240.255	Maya Pardazesh Iranian LLC
195.238.247.0	195.238.247.255	Web Gostaran Bandar Company (PJS)
195.245.70.0	195.245.70.255	Sanaye Khodrosazi Kerman (Kerman Car Industry) PLC
195.245.71.0	195.245.71.255	Sanaye Khodrosazi Kerman (Kerman Car Industry) PLC
204.18.0.0	204.18.0.255	Mobile Communication Company of Iran Plc
204.18.1.0	204.18.1.255	Mobile Communication Company of Iran Plc
204.18.2.0	204.18.3.255	Mobile Communication Company of Iran Plc
204.18.4.0	204.18.7.255	Mobile Communication Company of Iran Plc
204.18.8.0	204.18.15.255	Mobile Communication Company of Iran Plc
204.18.16.0	204.18.31.255	Mobile Communication Company of Iran Plc
204.18.32.0	204.18.47.255	Mobile Communication Company of Iran Plc
204.18.48.0	204.18.63.255	Mobile Communication Company of Iran Plc
204.18.64.0	204.18.79.255	Mobile Communication Company of Iran Plc
204.18.80.0	204.18.95.255	Mobile Communication Company of Iran Plc
204.18.96.0	204.18.111.255	Mobile Communication Company of Iran Plc
204.18.112.0	204.18.127.255	Mobile Communication Company of Iran Plc
204.18.128.0	204.18.128.255	Mobile Communication Company of Iran Plc
204.18.129.0	204.18.129.255	Mobile Communication Company of Iran Plc
204.18.130.0	204.18.131.255	Mobile Communication Company of Iran Plc
204.18.132.0	204.18.135.255	Mobile Communication Company of Iran Plc
204.18.136.0	204.18.143.255	Mobile Communication Company of Iran Plc
204.18.144.0	204.18.159.255	Mobile Communication Company of Iran Plc
204.18.160.0	204.18.175.255	Mobile Communication Company of Iran Plc
204.18.176.0	204.18.191.255	Mobile Communication Company of Iran Plc
204.18.192.0	204.18.207.255	Mobile Communication Company of Iran Plc
204.18.208.0	204.18.223.255	Mobile Communication Company of Iran Plc
204.18.224.0	204.18.239.255	Mobile Communication Company of Iran Plc
204.18.240.0	204.18.255.255	Mobile Communication Company of Iran Plc
212.1.192.0	212.1.192.255	Maskan Bank
212.1.193.0	212.1.193.255	Maskan Bank
212.1.194.0	212.1.194.255	Maskan Bank
212.1.195.0	212.1.195.255	Maskan Bank
212.1.196.0	212.1.199.255	Maskan Bank
212.16.64.0	212.16.64.255	Institute for Research in Fundamental Sciences
212.16.65.0	212.16.65.255	Institute for Research in Fundamental Sciences
212.16.66.0	212.16.66.255	Farhang Azma Communications Company LTD
212.16.67.0	212.16.67.255	Institute for Research in Fundamental Sciences
212.16.68.0	212.16.68.255	Farhang Azma Communications Company LTD
212.16.69.0	212.16.69.255	Farhang Azma Communications Company LTD
212.16.70.0	212.16.70.255	Farhang Azma Communications Company LTD
212.16.71.0	212.16.71.255	Sefroyek Pardaz Engineering PJSC
212.16.72.0	212.16.72.255	Farhang Azma Communications Company LTD
212.16.73.0	212.16.73.255	Farhang Azma Communications Company LTD
212.16.74.0	212.16.74.255	Farhang Azma Communications Company LTD
212.16.75.0	212.16.75.255	Institute for Research in Fundamental Sciences
212.16.76.0	212.16.76.255	Farhang Azma Communications Company LTD
212.16.77.0	212.16.77.255	Farhang Azma Communications Company LTD
212.16.79.0	212.16.79.255	Farhang Azma Communications Company LTD
212.16.81.0	212.16.81.255	Toesegaran Shabakeh Arseh Novin Ltd
212.16.82.0	212.16.82.255	Institute for Research in Fundamental Sciences
212.16.83.0	212.16.83.255	Institute for Research in Fundamental Sciences
212.16.84.0	212.16.84.255	Farhang Azma Communications Company LTD
212.16.86.0	212.16.86.255	Asre Pardazeshe Ettelaate Amin Institute
212.16.88.0	212.16.88.255	Institute for Research in Fundamental Sciences
212.16.89.0	212.16.89.255	PJSC Toseh Etemad Vosoogh Gostar
212.16.90.0	212.16.90.255	Institute for Research in Fundamental Sciences
212.16.91.0	212.16.91.255	Institute for Research in Fundamental Sciences
212.16.92.0	212.16.93.255	Farhang Azma Communications Company LTD
212.16.95.0	212.16.95.255	Farhang Azma Communications Company LTD
212.18.108.0	212.18.108.255	Faraso Samaneh Pasargad Ltd.
212.23.201.0	212.23.201.255	Vandad Vira Hooman LLC
212.23.214.0	212.23.214.255	Ali Monfared
212.23.216.0	212.23.216.255	Asiatech Data Transmission Company
212.33.192.0	212.33.193.255	Asiatech Data Transmission Company
212.33.194.0	212.33.194.255	Asiatech Data Transmission Company
212.33.195.0	212.33.195.255	Asiatech Data Transmission Company
212.33.196.0	212.33.196.255	Asiatech Data Transmission Company
212.33.197.0	212.33.197.255	Asiatech Data Transmission Company
212.33.198.0	212.33.198.255	Asiatech Data Transmission Company
212.33.199.0	212.33.199.255	Asiatech Data Transmission Company
212.33.200.0	212.33.200.255	Asiatech Data Transmission Company
212.33.201.0	212.33.201.255	Asiatech Data Transmission Company
212.33.202.0	212.33.202.255	Asiatech Data Transmission Company
212.33.203.0	212.33.203.255	Asiatech Data Transmission Company
212.33.204.0	212.33.204.255	Asiatech Data Transmission Company
212.33.205.0	212.33.205.255	Asiatech Data Transmission Company
212.33.206.0	212.33.206.255	Asiatech Data Transmission Company
212.33.207.0	212.33.207.255	Asiatech Data Transmission Company
212.33.208.0	212.33.209.255	Iran Telecommunication Company Pjs
212.33.210.0	212.33.211.255	Iran Telecommunication Company Pjs
212.33.212.0	212.33.213.255	Iran Telecommunication Company Pjs
212.33.214.0	212.33.215.255	Iran Telecommunication Company Pjs
212.33.216.0	212.33.223.255	Iran Telecommunication Company Pjs
212.46.45.0	212.46.45.255	Mapna Operation & Maintenace Company PJS
212.80.0.0	212.80.1.255	Farhang Azma Communications Company LTD
212.80.2.0	212.80.2.255	Farhang Azma Communications Company LTD
212.80.3.0	212.80.3.255	Institute for Research in Fundamental Sciences
212.80.4.0	212.80.4.255	Tehran Internet Co.
212.80.5.0	212.80.5.255	Institute for Research in Fundamental Sciences
212.80.8.0	212.80.8.255	Giti Secure Cloud LLC
212.80.9.0	212.80.9.255	Giti Secure Cloud LLC
212.80.10.0	212.80.10.255	Giti Secure Cloud LLC
212.80.11.0	212.80.11.255	Toesegaran Shabakeh Arseh Novin Ltd
212.80.12.0	212.80.12.255	Farhang Azma Communications Company LTD
212.80.13.0	212.80.13.255	Farhang Azma Communications Company LTD
212.80.14.0	212.80.14.255	Farhang Azma Communications Company LTD
212.80.15.0	212.80.15.255	Farhang Azma Communications Company LTD
212.80.16.0	212.80.16.255	Institute for Research in Fundamental Sciences
212.80.17.0	212.80.17.255	Farhang Azma Communications Company LTD
212.80.18.0	212.80.18.255	Asre Pardazeshe Ettelaate Amin Institute
212.80.19.0	212.80.19.255	Asre Pardazeshe Ettelaate Amin Institute
212.80.20.0	212.80.20.255	Asre Pardazeshe Ettelaate Amin Institute
212.80.21.0	212.80.21.255	Asre Pardazeshe Ettelaate Amin Institute
212.80.22.0	212.80.22.255	Asre Pardazeshe Ettelaate Amin Institute
212.80.23.0	212.80.23.255	Asre Pardazeshe Ettelaate Amin Institute
212.80.24.0	212.80.24.255	Parmin Cloud Computing LLC
212.80.25.0	212.80.25.255	Parsian Electronic Commerce PJSC
212.80.26.0	212.80.26.255	Parsian Electronic Commerce PJSC
212.80.27.0	212.80.27.255	Parsian Electronic Commerce PJSC
212.80.28.0	212.80.28.255	Farhang Azma Communications Company LTD
212.80.29.0	212.80.29.255	Peyman Ertebatat Pouya Company (Ltd)
212.80.30.0	212.80.30.255	Institute for Research in Fundamental Sciences
212.80.31.0	212.80.31.255	Sefroyek Pardaz Engineering PJSC
212.86.64.0	212.86.64.255	Homaye Jahan Nama Co. ( Private Joint Stock)
212.86.65.0	212.86.65.255	Homaye Jahan Nama Co. ( Private Joint Stock)
212.86.66.0	212.86.66.255	Homaye Jahan Nama Co. ( Private Joint Stock)
212.86.67.0	212.86.67.255	Homaye Jahan Nama Co. ( Private Joint Stock)
212.86.68.0	212.86.68.255	Homaye Jahan Nama Co. ( Private Joint Stock)
212.86.69.0	212.86.69.255	Homaye Jahan Nama Co. ( Private Joint Stock)
212.86.70.0	212.86.70.255	Homaye Jahan Nama Co. ( Private Joint Stock)
212.86.71.0	212.86.71.255	Homaye Jahan Nama Co. ( Private Joint Stock)
212.86.72.0	212.86.72.255	Homaye Jahan Nama Co. ( Private Joint Stock)
212.86.73.0	212.86.73.255	Homaye Jahan Nama Co. ( Private Joint Stock)
212.86.74.0	212.86.74.255	Homaye Jahan Nama Co. ( Private Joint Stock)
212.86.75.0	212.86.75.255	Homaye Jahan Nama Co. ( Private Joint Stock)
212.86.76.0	212.86.76.255	Homaye Jahan Nama Co. ( Private Joint Stock)
212.86.77.0	212.86.77.255	Homaye Jahan Nama Co. ( Private Joint Stock)
212.86.78.0	212.86.78.255	Homaye Jahan Nama Co. ( Private Joint Stock)
212.86.79.0	212.86.79.255	Homaye Jahan Nama Co. ( Private Joint Stock)
212.86.80.0	212.86.80.255	Homaye Jahan Nama Co. ( Private Joint Stock)
212.86.81.0	212.86.81.255	Homaye Jahan Nama Co. ( Private Joint Stock)
212.86.82.0	212.86.82.255	Homaye Jahan Nama Co. ( Private Joint Stock)
212.86.83.0	212.86.83.255	Homaye Jahan Nama Co. ( Private Joint Stock)
212.86.84.0	212.86.84.255	Homaye Jahan Nama Co. ( Private Joint Stock)
212.86.85.0	212.86.85.255	Homaye Jahan Nama Co. ( Private Joint Stock)
212.86.86.0	212.86.86.255	Homaye Jahan Nama Co. ( Private Joint Stock)
212.86.87.0	212.86.87.255	Homaye Jahan Nama Co. ( Private Joint Stock)
212.86.88.0	212.86.88.255	Homaye Jahan Nama Co. ( Private Joint Stock)
212.86.89.0	212.86.89.255	Homaye Jahan Nama Co. ( Private Joint Stock)
212.86.90.0	212.86.90.255	Homaye Jahan Nama Co. ( Private Joint Stock)
212.86.91.0	212.86.91.255	Homaye Jahan Nama Co. ( Private Joint Stock)
212.86.92.0	212.86.92.255	Homaye Jahan Nama Co. ( Private Joint Stock)
212.86.93.0	212.86.93.255	Homaye Jahan Nama Co. ( Private Joint Stock)
212.86.94.0	212.86.94.255	Homaye Jahan Nama Co. ( Private Joint Stock)
212.86.95.0	212.86.95.255	Homaye Jahan Nama Co. ( Private Joint Stock)
212.108.98.0	212.108.98.255	Giti Secure Cloud LLC
212.108.102.0	212.108.102.255	Sorat Gostar Abr Datis LLC
212.108.125.0	212.108.125.255	Pooya Parto Qeshm Cooperative Company
212.108.127.0	212.108.127.255	Amir Feizi
212.120.192.0	212.120.195.255	Hamara System Tabriz Engineering Company
212.120.196.0	212.120.197.255	Hamara System Tabriz Engineering Company
212.120.198.0	212.120.199.255	Hamara System Tabriz Engineering Company
212.120.200.0	212.120.207.255	Hamara System Tabriz Engineering Company
212.120.208.0	212.120.223.255	Hamara System Tabriz Engineering Company
213.108.240.0	213.108.240.255	Dadeh Gostar Sina Shiraz PJSC
213.108.241.0	213.108.241.255	Dadeh Gostar Sina Shiraz PJSC
213.108.242.0	213.108.242.255	Dadeh Gostar Sina Shiraz PJSC
213.108.243.0	213.108.243.255	Dadeh Gostar Sina Shiraz PJSC
213.109.199.0	213.109.199.255	Asiatech Data Transmission Company
213.109.240.0	213.109.247.255	Iran Telecommunication Company Pjs
213.109.248.0	213.109.255.255	Iran Telecommunication Company Pjs
213.176.1.0	213.176.1.255	Iranian Research Organization for Science & Technology
213.176.2.0	213.176.2.255	Mohammadreza Gholami
213.176.3.0	213.176.3.255	Iranian Research Organization for Science & Technology
213.176.4.0	213.176.4.255	Iranian Research Organization for Science & Technology
213.176.5.0	213.176.5.255	Iranian Research Organization for Science & Technology
213.176.6.0	213.176.6.255	Ettelaat Fanavarn-E Tookan Co Ltd
213.176.7.0	213.176.7.255	Enteghal Dadeh Mahan Co. PJSC
213.176.20.0	213.176.23.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
213.176.24.0	213.176.27.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
213.176.28.0	213.176.28.255	Iranian Research Organization for Science & Technology
213.176.29.0	213.176.29.255	Noavaran Shabakeh Sabz Mehregan (Ltd.)
213.176.30.0	213.176.30.255	Noavaran Shabakeh Sabz Mehregan (Ltd.)
213.176.31.0	213.176.31.255	Noavaran Shabakeh Sabz Mehregan (Ltd.)
213.176.68.0	213.176.71.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
213.176.88.0	213.176.91.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
213.176.120.0	213.176.120.255	Parsway Shomal Company Ltd
213.176.121.0	213.176.121.255	Iranian Research Organization for Science & Technology
213.176.122.0	213.176.122.255	Iranian Research Organization for Science & Technology
213.176.123.0	213.176.123.255	Iranian Research Organization for Science & Technology
213.176.124.0	213.176.124.255	Iranian Research Organization for Science & Technology
213.176.125.0	213.176.125.255	Iranian Research Organization for Science & Technology
213.176.126.0	213.176.126.255	Iranian Research Organization for Science & Technology
213.176.127.0	213.176.127.255	Enteghal Dadeh Mahan Co. PJSC
213.177.176.0	213.177.176.255	Taymaz Gostaran Naghshe Almas Co. ( Ltd )
213.177.178.0	213.177.178.255	Maya Pardazesh Iranian LLC
213.177.181.0	213.177.181.255	Isfahan Science & Technology Town
213.195.0.0	213.195.7.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
213.195.8.0	213.195.15.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
213.195.16.0	213.195.19.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
213.195.20.0	213.195.20.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
213.195.21.0	213.195.21.255	Bank Pasargad PJSC
213.195.22.0	213.195.22.255	Rasaneh Avabarid Private Joint Stock Company
213.195.23.0	213.195.23.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
213.195.32.0	213.195.35.255	Rightel Communication Service Company PJS
213.195.36.0	213.195.37.255	Rightel Communication Service Company PJS
213.195.38.0	213.195.39.255	Rightel Communication Service Company PJS
213.195.40.0	213.195.47.255	Rightel Communication Service Company PJS
213.195.48.0	213.195.51.255	Rightel Communication Service Company PJS
213.195.52.0	213.195.55.255	Iran Telecommunication Company Pjs
213.195.56.0	213.195.63.255	Iran Telecommunication Company Pjs
213.207.192.0	213.207.193.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
213.207.194.0	213.207.194.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
213.207.195.0	213.207.195.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
213.207.196.0	213.207.196.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
213.207.197.0	213.207.197.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
213.207.198.0	213.207.198.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
213.207.199.0	213.207.199.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
213.207.200.0	213.207.200.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
213.207.201.0	213.207.201.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
213.207.202.0	213.207.202.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
213.207.203.0	213.207.203.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
213.207.204.0	213.207.204.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
213.207.205.0	213.207.205.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
213.207.206.0	213.207.207.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
213.207.208.0	213.207.223.255	Datak Company LLC
213.207.224.0	213.207.231.255	Kish Cell Pars Co. (PJS)
213.207.232.0	213.207.239.255	Kish Cell Pars Co. (PJS)
213.207.240.0	213.207.255.255	Datak Company LLC
213.232.124.0	213.232.124.255	Toesegaran Shabakeh Arseh Novin Ltd
213.232.125.0	213.232.125.255	Toesegaran Shabakeh Arseh Novin Ltd
213.232.126.0	213.232.126.255	Toesegaran Shabakeh Arseh Novin Ltd
213.232.127.0	213.232.127.255	Toesegaran Shabakeh Arseh Novin Ltd
213.233.160.0	213.233.160.255	Sharif University of Technology
213.233.161.0	213.233.161.255	Sharif University of Technology
213.233.162.0	213.233.162.255	Sharif University of Technology
213.233.163.0	213.233.163.255	Sharif University of Technology
213.233.164.0	213.233.164.255	Sharif University of Technology
213.233.165.0	213.233.165.255	Sharif University of Technology
213.233.166.0	213.233.167.255	Sharif University of Technology
213.233.168.0	213.233.171.255	Sharif University of Technology
213.233.172.0	213.233.172.255	Sharif University of Technology
213.233.173.0	213.233.173.255	Sharif University of Technology
213.233.174.0	213.233.175.255	Sharif University of Technology
213.233.176.0	213.233.183.255	Sharif University of Technology
213.233.184.0	213.233.184.255	Sharif University of Technology
213.233.185.0	213.233.185.255	Sharif University of Technology
213.233.186.0	213.233.187.255	Sharif University of Technology
213.233.188.0	213.233.191.255	Sharif University of Technology
217.11.16.0	217.11.16.255	Afranet
217.11.17.0	217.11.17.255	Afranet
217.11.18.0	217.11.18.255	Afranet
217.11.19.0	217.11.19.255	Afranet
217.11.20.0	217.11.20.255	Afranet
217.11.21.0	217.11.21.255	Afranet
217.11.22.0	217.11.22.255	Afranet
217.11.23.0	217.11.23.255	Afranet
217.11.24.0	217.11.24.255	Afranet
217.11.25.0	217.11.25.255	Afranet
217.11.26.0	217.11.26.255	Afranet
217.11.27.0	217.11.27.255	Afranet
217.11.28.0	217.11.28.255	Afranet
217.11.29.0	217.11.29.255	Afranet
217.11.30.0	217.11.30.255	Afranet
217.11.31.0	217.11.31.255	Afranet
217.18.48.0	217.18.48.255	Parmin Cloud Computing LLC
217.18.90.0	217.18.90.255	Patron Technology Persia Ltd
217.18.94.0	217.18.94.255	Taymaz Gostaran Naghshe Almas Co. ( Ltd )
217.20.252.0	217.20.252.255	Dadeh Gostaran Webideh Co. Ltd.
217.24.144.0	217.24.159.255	Iran Telecommunication Company Pjs
217.25.48.0	217.25.48.255	IRNA
217.25.49.0	217.25.49.255	IRNA
217.25.50.0	217.25.50.255	IRNA
217.25.51.0	217.25.51.255	IRNA
217.25.52.0	217.25.52.255	IRNA
217.25.53.0	217.25.53.255	IRNA
217.25.54.0	217.25.54.255	IRNA
217.25.55.0	217.25.55.255	IRNA
217.25.56.0	217.25.56.255	IRNA
217.25.57.0	217.25.57.255	IRNA
217.25.58.0	217.25.58.255	IRNA
217.25.59.0	217.25.59.255	IRNA
217.25.60.0	217.25.60.255	IRNA
217.25.61.0	217.25.61.255	IRNA
217.25.62.0	217.25.62.255	IRNA
217.25.63.0	217.25.63.255	IRNA
217.26.222.0	217.26.222.255	Fanavaran Nicsepehr Zenderood Ltd.
217.60.255.0	217.60.255.255	Atrin Information & Communications Technology Company PJS
217.66.192.0	217.66.195.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
217.66.196.0	217.66.197.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
217.66.198.0	217.66.198.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
217.66.199.0	217.66.199.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
217.66.200.0	217.66.200.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
217.66.201.0	217.66.201.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
217.66.202.0	217.66.202.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
217.66.203.0	217.66.203.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
217.66.204.0	217.66.204.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
217.66.205.0	217.66.205.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
217.66.206.0	217.66.207.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
217.66.208.0	217.66.208.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
217.66.209.0	217.66.209.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
217.66.210.0	217.66.210.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
217.66.211.0	217.66.211.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
217.66.212.0	217.66.212.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
217.66.213.0	217.66.213.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
217.66.214.0	217.66.215.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
217.66.216.0	217.66.216.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
217.66.217.0	217.66.217.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
217.66.218.0	217.66.218.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
217.66.219.0	217.66.219.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
217.66.220.0	217.66.223.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
217.77.112.0	217.77.119.255	Iran Telecommunication Company Pjs
217.77.120.0	217.77.127.255	Iran Telecommunication Company Pjs
217.114.40.0	217.114.40.255	Vandad Vira Hooman LLC
217.114.46.0	217.114.46.255	MEHRAN_HEIDARI
217.144.104.0	217.144.104.255	Netmihan Communication Company Ltd
217.144.105.0	217.144.105.255	Netmihan Communication Company Ltd
217.144.106.0	217.144.106.255	Netmihan Communication Company Ltd
217.144.107.0	217.144.107.255	Netmihan Communication Company Ltd
217.146.208.0	217.146.211.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
217.146.212.0	217.146.213.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
217.146.214.0	217.146.214.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
217.146.215.0	217.146.215.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
217.146.216.0	217.146.223.255	Tose'H Fanavari Ertebabat Pasargad Arian Co. PJS
217.170.240.0	217.170.241.255	Petiak System Co JSC
217.170.242.0	217.170.243.255	Petiak System Co JSC
217.170.244.0	217.170.244.255	Petiak System Co JSC
217.170.245.0	217.170.245.255	Petiak System Co JSC
217.170.246.0	217.170.247.255	Petiak System Co JSC
217.170.248.0	217.170.249.255	Petiak System Co JSC
217.170.250.0	217.170.251.255	Petiak System Co JSC
217.170.252.0	217.170.253.255	Petiak System Co JSC
217.170.254.0	217.170.255.255	Petiak System Co JSC
217.171.145.0	217.171.145.255	Iran Telecommunication Company Pjs
217.171.148.0	217.171.151.255	Iran Telecommunication Company Pjs
217.172.96.0	217.172.97.255	Moshaverin Parsaye Barin Company Ltd.
217.172.98.0	217.172.98.255	Asiatech Data Transmission Company
217.172.99.0	217.172.99.255	Moshaverin Parsaye Barin Company Ltd.
217.172.100.0	217.172.100.255	Moshaverin Parsaye Barin Company Ltd.
217.172.101.0	217.172.101.255	Moshaverin Parsaye Barin Company Ltd.
217.172.102.0	217.172.103.255	Iran Telecommunication Company Pjs
217.172.104.0	217.172.111.255	Iran Telecommunication Company Pjs
217.172.112.0	217.172.113.255	Iran Telecommunication Company Pjs
217.172.114.0	217.172.115.255	Iran Telecommunication Company Pjs
217.172.116.0	217.172.117.255	Iran Telecommunication Company Pjs
217.172.118.0	217.172.118.255	Iran Telecommunication Company Pjs
217.172.119.0	217.172.119.255	Iran Telecommunication Company Pjs
217.172.120.0	217.172.127.255	Sefroyek Pardaz Engineering PJSC
217.174.16.0	217.174.31.255	National Iranian Oil Company
217.198.190.0	217.198.190.255	Fanavaran Nicsepehr Zenderood Ltd.
217.218.0.0	217.218.3.255	Iran Telecommunication Company Pjs
217.218.4.0	217.218.5.255	Iran Telecommunication Company Pjs
217.218.6.0	217.218.6.255	Iran Telecommunication Company Pjs
217.218.7.0	217.218.7.255	Iran Telecommunication Company Pjs
217.218.8.0	217.218.11.255	Iran Telecommunication Company Pjs
217.218.12.0	217.218.13.255	Iran Telecommunication Company Pjs
217.218.14.0	217.218.14.255	Iran Telecommunication Company Pjs
217.218.15.0	217.218.15.255	Iran Telecommunication Company Pjs
217.218.16.0	217.218.31.255	Iran Telecommunication Company Pjs
217.218.32.0	217.218.47.255	Iran Telecommunication Company Pjs
217.218.48.0	217.218.51.255	Iran Telecommunication Company Pjs
217.218.52.0	217.218.52.255	Iran Telecommunication Company Pjs
217.218.53.0	217.218.53.255	Iran Telecommunication Company Pjs
217.218.54.0	217.218.55.255	Iran Telecommunication Company Pjs
217.218.56.0	217.218.63.255	Iran Telecommunication Company Pjs
217.218.64.0	217.218.65.255	Telecommunication Infrastructure Company
217.218.66.0	217.218.66.255	Iran Telecommunication Company Pjs
217.218.67.0	217.218.67.255	Telecommunication Infrastructure Company
217.218.68.0	217.218.71.255	Iran Telecommunication Company Pjs
217.218.72.0	217.218.79.255	Iran Telecommunication Company Pjs
217.218.80.0	217.218.95.255	Iran Telecommunication Company Pjs
217.218.96.0	217.218.111.255	Iran Telecommunication Company Pjs
217.218.112.0	217.218.119.255	Iran Telecommunication Company Pjs
217.218.120.0	217.218.123.255	Iran Telecommunication Company Pjs
217.218.124.0	217.218.125.255	Iran Telecommunication Company Pjs
217.218.126.0	217.218.126.255	Iran Telecommunication Company Pjs
217.218.127.0	217.218.127.255	Iran Information Technology Company PJSC
217.218.128.0	217.218.143.255	Iran Telecommunication Company Pjs
217.218.144.0	217.218.151.255	Iran Telecommunication Company Pjs
217.218.152.0	217.218.153.255	Iran Telecommunication Company Pjs
217.218.154.0	217.218.154.255	Iran Telecommunication Company Pjs
217.218.155.0	217.218.155.255	Iran Information Technology Company PJSC
217.218.156.0	217.218.156.255	Iran Telecommunication Company Pjs
217.218.157.0	217.218.157.255	Iran Telecommunication Company Pjs
217.218.158.0	217.218.159.255	Iran Telecommunication Company Pjs
217.218.160.0	217.218.163.255	Iran Telecommunication Company Pjs
217.218.164.0	217.218.164.255	Iran Telecommunication Company Pjs
217.218.165.0	217.218.165.255	Iran Information Technology Company PJSC
217.218.166.0	217.218.167.255	Iran Telecommunication Company Pjs
217.218.168.0	217.218.175.255	Iran Telecommunication Company Pjs
217.218.176.0	217.218.177.255	Iran Telecommunication Company Pjs
217.218.178.0	217.218.178.255	Iran Telecommunication Company Pjs
217.218.179.0	217.218.179.255	Iran Telecommunication Company Pjs
217.218.180.0	217.218.180.255	Iran Telecommunication Company Pjs
217.218.181.0	217.218.181.255	Iran Telecommunication Company Pjs
217.218.182.0	217.218.182.255	Iran Telecommunication Company Pjs
217.218.183.0	217.218.183.255	Iran Telecommunication Company Pjs
217.218.184.0	217.218.185.255	Telecommunication Infrastructure Company
217.218.186.0	217.218.187.255	Iran Telecommunication Company Pjs
217.218.188.0	217.218.191.255	Iran Telecommunication Company Pjs
217.218.192.0	217.218.195.255	Iran Telecommunication Company Pjs
217.218.196.0	217.218.197.255	Iran Telecommunication Company Pjs
217.218.198.0	217.218.199.255	Iran Telecommunication Company Pjs
217.218.200.0	217.218.201.255	Iran Telecommunication Company Pjs
217.218.202.0	217.218.203.255	Iran Telecommunication Company Pjs
217.218.204.0	217.218.205.255	Iran Telecommunication Company Pjs
217.218.206.0	217.218.206.255	Iran Telecommunication Company Pjs
217.218.207.0	217.218.207.255	Iran Telecommunication Company Pjs
217.218.208.0	217.218.215.255	Iran Telecommunication Company Pjs
217.218.216.0	217.218.217.255	Iran Telecommunication Company Pjs
217.218.218.0	217.218.219.255	Iran Telecommunication Company Pjs
217.218.220.0	217.218.223.255	Iran Telecommunication Company Pjs
217.218.224.0	217.218.231.255	Iran Telecommunication Company Pjs
217.218.232.0	217.218.239.255	Iran Telecommunication Company Pjs
217.218.240.0	217.218.247.255	Iran Telecommunication Company Pjs
217.218.248.0	217.218.255.255	Iran Telecommunication Company Pjs
217.219.0.0	217.219.15.255	Iran Telecommunication Company Pjs
217.219.16.0	217.219.19.255	Iran Telecommunication Company Pjs
217.219.20.0	217.219.23.255	Iran Telecommunication Company Pjs
217.219.24.0	217.219.31.255	Iran Telecommunication Company Pjs
217.219.32.0	217.219.39.255	Iran Telecommunication Company Pjs
217.219.40.0	217.219.43.255	Iran Telecommunication Company Pjs
217.219.44.0	217.219.47.255	Iran Telecommunication Company Pjs
217.219.48.0	217.219.51.255	Iran Telecommunication Company Pjs
217.219.52.0	217.219.55.255	Iran Telecommunication Company Pjs
217.219.56.0	217.219.63.255	Iran Telecommunication Company Pjs
217.219.64.0	217.219.65.255	Iran Telecommunication Company Pjs
217.219.66.0	217.219.66.255	Iran Telecommunication Company Pjs
217.219.67.0	217.219.67.255	Iran Telecommunication Company Pjs
217.219.68.0	217.219.71.255	Iran Telecommunication Company Pjs
217.219.72.0	217.219.75.255	Iran Telecommunication Company Pjs
217.219.76.0	217.219.77.255	Iran Telecommunication Company Pjs
217.219.78.0	217.219.78.255	Iran Telecommunication Company Pjs
217.219.79.0	217.219.79.255	Iran Telecommunication Company Pjs
217.219.80.0	217.219.87.255	Iran Telecommunication Company Pjs
217.219.88.0	217.219.88.255	Iran Telecommunication Company Pjs
217.219.89.0	217.219.89.255	Iran Telecommunication Company Pjs
217.219.90.0	217.219.91.255	Iran Telecommunication Company Pjs
217.219.92.0	217.219.95.255	Iran Telecommunication Company Pjs
217.219.96.0	217.219.111.255	Iran Telecommunication Company Pjs
217.219.112.0	217.219.119.255	Iran Telecommunication Company Pjs
217.219.120.0	217.219.127.255	Iran Telecommunication Company Pjs
217.219.128.0	217.219.143.255	Iran Telecommunication Company Pjs
217.219.144.0	217.219.151.255	Iran Telecommunication Company Pjs
217.219.152.0	217.219.159.255	Iran Telecommunication Company Pjs
217.219.160.0	217.219.175.255	Iran Telecommunication Company Pjs
217.219.176.0	217.219.183.255	Iran Telecommunication Company Pjs
217.219.184.0	217.219.191.255	Iran Telecommunication Company Pjs
217.219.192.0	217.219.199.255	Iran Telecommunication Company Pjs
217.219.200.0	217.219.207.255	Iran Telecommunication Company Pjs
217.219.208.0	217.219.211.255	Iran Telecommunication Company Pjs
217.219.212.0	217.219.215.255	Iran Telecommunication Company Pjs
217.219.216.0	217.219.219.255	Iran Telecommunication Company Pjs
217.219.220.0	217.219.223.255	Iran Telecommunication Company Pjs
217.219.224.0	217.219.231.255	Iran Telecommunication Company Pjs
217.219.232.0	217.219.239.255	Iran Telecommunication Company Pjs
217.219.240.0	217.219.243.255	Iran Telecommunication Company Pjs
217.219.244.0	217.219.247.255	Iran Telecommunication Company Pjs
217.219.248.0	217.219.249.255	Iran Telecommunication Company Pjs
217.219.250.0	217.219.251.255	Iran Telecommunication Company Pjs
217.219.252.0	217.219.253.255	Iran Telecommunication Company Pjs
217.219.254.0	217.219.254.255	Iran Telecommunication Company Pjs
217.219.255.0	217.219.255.255	Telecommunication Infrastructure Company
`

func init() {
	lines := strings.Split(strings.TrimSpace(rawIRRanges), "\n")
	irRanges = make([]irRange, 0, len(lines))
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l == "" {
			continue
		}
		parts := strings.Split(l, "\t")
		if len(parts) < 3 {
			continue
		}
		from := net.ParseIP(parts[0])
		to := net.ParseIP(parts[1])
		if from != nil && to != nil {
			irRanges = append(irRanges, irRange{
				from: from,
				to:   to,
				isp:  parts[2],
			})
		}
	}
}

// LookupIranISP checks if the given IP address falls within any of the static Iranian ISP IP blocks.
// If it does, it returns the ISP owner's name and true.
func LookupIranISP(ipStr string) (string, bool) {
	ip := net.ParseIP(strings.TrimSpace(ipStr))
	if ip == nil {
		return "", false
	}
	ip16 := ip.To16()
	if ip16 == nil {
		return "", false
	}
	for _, r := range irRanges {
		from16 := r.from.To16()
		to16 := r.to.To16()
		if from16 != nil && to16 != nil {
			if bytes.Compare(ip16, from16) >= 0 && bytes.Compare(ip16, to16) <= 0 {
				return r.isp, true
			}
		}
	}
	return "", false
}
