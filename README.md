# RandomPictureAPI
A simple API returns a random picture

## Config
see config.toml.example

## API
### API Authorization
Access path of /api/* needs authorization.  
Set api key into http header: Authorization.
### Artists List
```
GET /api/artists
```
Returns  
```
{
    name: string,
    pid: string,
    count: number
}[]
```
### Random Picture
#### 1.Every artist has the same probability
```
GET /api/random/artists
```
#### 2.Every picture has the same probability
```
GET /api/random/pics
```
#### 3.Get a random picture of an artist
```
GET /api/random/artists/:pid
```
3 APIs above returns  
```
{
    pic_req_id: string,
    artist_info: {
        name: string,
        pid: string,
        count: number
    }
}
```
pic_req_id valid within five minutes

## Get Random Picture
```
GET /image/:pic_req_id
```
Returns picture file  
This path does not need authorization