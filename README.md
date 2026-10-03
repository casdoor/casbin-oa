# Casbin-OA

Casbin-OA is An official manuscript processing, evaluation and display system for Casbin technical writers

## Online demo

Deployed site: https://oa.casbin.com/

## Architecture

Casbin-oa contains 2 parts:

Name | Description | Language | Source code
----|------|----|----
Frontend | Web frontend UI for Casbin-oa | Javascript + React | https://github.com/casbin/casbin-oa/tree/master/web
Backend | RESTful API backend for Casbin-oa | Golang + Beego + MySQL | https://github.com/casbin/casbin-oa/


## Installation

Casbin-OA uses Casdoor to manage members. So you need to create an organization and an application for Casnode in a
Casdoor instance.

### Necessary configuration

- ##### Get the code:

  ```shell
  go get github.com/casdoor/casdoor
  go get github.com/casbin/casbin-oa
  ```
  or
    ```shell
    git clone https://github.com/casdoor/casdoor
    git clone https://github.com/casbin/casbin-oa.git
    ```

- Setup database:

  Casbin-oa will store its student, report and topics informations in a MySQL database named: `casbin_oa`, will create
  it if not existed. The DB connection string can be specified at: https://github.com/casbin/casbin-oa/tree/master/conf

    ```ini
    dataSourceName = root:123@tcp(localhost:3306)/
    ```

  Casbin-oa uses XORM to connect to DB, so all DBs supported by XORM can also be used.

- Run backend (in port 10000):

    ```shell
    go run main.go
    ```

- Run frontend (in the same machine's port 9000):

    ```shell
    cd web
    ## npm
    npm install
    npm run start
    ## yarn
    yarn install
    yarn run start
    ```

- Open browser:

  http://localhost:9000/

### Optional configuration

#### Setup your forum to enable some third-party login platform

Casbin-OA uses Casdoor to manage members. If you want to log in with oauth, you should
see [casdoor oauth configuration](https://casdoor.ai/docs/provider/oauth/overview/).

#### OSS, Mail, and SMS services

Casbin-OA uses Casdoor to upload files to cloud storage, send Emails and send SMSs. See Casdoor for more details.

#### Github corner

We added a Github icon in the upper right corner, linking to your Github repository address. You could
set `ShowGithubCorner` to hidden it.

Configuration:

```
export const ShowGithubCorner = true

export const GithubRepo = "https://github.com/casbin/casbin-oa" //your github repository
```

