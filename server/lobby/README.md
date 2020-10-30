WSNet2 Lobby API
================

## Create Room

ednpoint
: POST /rooms

### エラーレスポンス
| 概要 | HTTP Status | gRPC Code | 発生箇所  | 備考 |
|------|-------------|-----------|-----------|------|
| レスポンスのmsgpackエンコード失敗 | InternalServerError | - | lobby/service/api.go: renderResponse() | - |
| ユーザ認証失敗 | Unauthorized | - | lobby/service/api.go: LobbyService.authUser() | - |
| リクエストbodyのmsgpackデコード失敗 | BadRequest | - | lobby/service/api.go: handleCreateRoom() | - |
| appIdのAppが無い | InternalServerError | - | lobby/room.go: RoomService.Create() | ユーザ認証失敗しているはずなので起こらない |
| gameサーバ取得失敗 | InternalServerError | - | lobby/game_cache.go: GameCache.Rand() | 生きているgameが見つからない |
| gRPC ClientをPoolから取得失敗 | InternalServerError | - | lobby/room.go: RoomService.Create() | - |
| gRPCタイムアウト | InternalServerError | DeadlineExceeded | lobby/room.go: RoomService.Create() | lobby側で設定したタイムアウト |
| appIdのAppが無い | InternalServerError | Internal | game/service/grpc.go: GameService.Create() | ユーザ認証失敗しているはずなので起こらない |
| context timeout | InternalServerError | DeadlineExceeded | game/repository.go: Repository.CreateRoom(), game/room.go: NewRoom() | game側で設定したタイムアウト |
| roomレコード作成失敗 | InternalServerError | Internal | game/repository.go: Repository.newRoomInfo() | - |
| room数上限 | ServiceUnavailable | ResourceExhausted | game/repository.go: Repository.CreateRoom() | - |
| DB Commit失敗 | InternalServerError | Internal | game/repository.go: Repository.CreateRoom() | - |
| {public,private}PropsのUnmarshal失敗| BadRequest | InvalidArgument | game/room.go: NewRoom() | - |
| Player PropsのUnmarshal失敗 | BadRequest | InvalidArgument | game/client.go: newClient() | - |


## Join Room

POST /rooms/join/id/{roomId}
POST /rooms/join/number/{roomNumber}

### エラーレスポンス
| 概要 | HTTP Status | gRPC Code | 発生箇所  | 備考 |
|------|-------------|-----------|-----------|------|
| レスポンスのmsgpackエンコード失敗 | InternalServerError | - | lobby/service/api.go: renderResponse() | - |
| ユーザ認証失敗 | Unauthorized | - | lobby/service/api.go: LobbyService.authUser() | - |
| リクエストbodyのmsgpackデコード失敗 | BadRequest | - | lobby/service/api.go: handleCreateRoom() | - |
| RoomIDが空 | BadRequest | - | lobby/service/api.go: handleJoinRoom() | - |
| RoomNumberが空 | BadRequest | - | lobby/service/api.go: handleJoinRoomByNumber() | - |
| appIdのAppが無い | InternalServerError | - | lobby/room.go: RoomService.JoinBy{Id,Number}() | ユーザ認証失敗しているはずなので起こらない |
| Roomが見つからない | **200 OK** | - | lobby/room.go: RoomService.JoinBy{Id,Number}() | - |
| プロパティクエリ条件に合致しない | **200 OK** | - | lobby/room.go: RoomService.JoinBy{Id,Number}() | - |
| publicPropsのデコード失敗 | InternalServerError | - | obby/room.go: RoomService.JoinBy{Id,Number}() | - |
| gameサーバ取得失敗 | InternalServerError | - | lobby/game_cache.go: GameCache.Get() | - |
| gRPC ClientをPoolから取得失敗 | InternalServerError | - | lobby/room.go: RoomService.join() | - |
| gRPCタイムアウト | InternalServerError | DeadlineExceeded | lobby/room.go: RoomService.join() | lobby側で設定したタイムアウト |
| appIdのAppが無い | InternalServerError | Internal | game/service/grpc.go: GameService.Join() | ユーザ認証失敗しているはずなので起こらない |
| gRPCタイムアウト | InternalServerError | Deadlineexceeded | game/repository.go: Repository.joinRoom() | game側で設定したタイムアウト |
| Roomが既に消えた | **200 OK** | NotFound | game/repository.go: Repository.joinRoom() | lobbyでのチェック後に消えたパターン |
| Joinableでない | **200 OK** | FailedPrecondition | game/room.go: msgJoin() | lobbyでのチェック後に折られた |
| 既に入室済み | Conflict | AlreadyExists | game/room.go: msgJoin() | - |
| 満室 | **200 OK** | ResourceExhausted | game/room.go: msgJoin() | - |
| Player PropsのUnmarshal失敗 | BadRequest | InvalidArgument | game/client.go: newClient() | - |


## Random Join

POST /rooms/join/random/{searchGroup}

### エラーレスポンス
| 概要 | HTTP Status | gRPC Code | 発生箇所  | 備考 |
|------|-------------|-----------|-----------|------|
| レスポンスのmsgpackエンコード失敗 | InternalServerError | - | lobby/service/api.go: renderResponse() | - |
| ユーザ認証失敗 | Unauthorized | - | lobby/service/api.go: LobbyService.authUser() | - |
| リクエストbodyのmsgpackデコード失敗 | BadRequest | - | lobby/service/api.go: handleJoinAtRandom() | - |
| GameCacheからの取得失敗 | InternalServerError | - | lobby/room_cache.go: roomCacheQuery.do() | - |
| 入室可能な部屋がない | **200 OK** | - | lobby/room.go: JoinAtRandom() | - |

