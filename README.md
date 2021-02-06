Deployed link: https://outstagram-be.herokuapp.com

# Image Repository (backend)
A simple image repository backend project built in Golang. This was my first time writing an application in Go and with Docker, and it was a great experience :)

Name is a pun on Instagram.

## Basic app architecture

This app has a Golang backend that will processes all API requests; it connects to a MongoDB database that stores user and image metadata.

The images are stored as static files in an Amazon S3 bucket, and the app is built in a Docker container.

Authentication is handled via JWTs issued by the server.

## Setup
1. Clone the repository.
2. Fill out the variables as required in `.env.example`. You will need the following:
    - MongoDB instance (I use Atlas)
    - AWS S3 keys
    - A key to generate your JWTs
3. Run the `Dockerfile`.
4. Done! You should be able to run the backend server.

## API information
Endpoint structure:

- `/api/users` routes to manage user authentication. Most operations require that users authenticate via a JWT.
- `/api/images` routes that contain operations related to images. Some routes require authentication; some don't.

## List of endpoints

### /users endpoints

#### [POST] /signup

**Accepts**: `application/x-www-form-urlencoded`

Handles sign up. Takes in a name, email, userHandle and password, verifies the inputs, and creates the user.

Passwords are subject to complexity requirements of:
- At least 8 characters
- 1 Uppercase
- 1 Lowercase
- 1 Digit

User Handles must be alphanumeric only.

##### Form fields:
- `email`: User's email address.
- `password`: User's password.
- `name`: User's given name.
- `userHandle`: User's desired user handle.

##### Returns:
- `200`: User was created. Returns ID.
- `400`: If any complexity requirement was not met, an invalid email was sent, or an invalid form was sent.
- `409`: Conflict, if the user was already registered with the email or userHandle.
- `500`: If there is an error on the server (Database error, etc).
___

#### [POST] /login
**Accepts**: `application/x-www-form-urlencoded`

Handles a login request.

##### Form fields:
- `email`: User's email address.
- `password`: User's password.

##### Returns:
- `200`: OK, if the username and password match. Will return the userinfo, and set userinfo and JWT cookies (that expire after an hour)
- `404`: If the user was not found, or the username and password don't match. (there is no difference here.)
- `500`: If there is an internal server error.
___
#### [GET] /getUsers

Gets users that match the query (from querystring). An empty query will return the first 100 users.

Can be used to search for users based on the given criteria.

##### Accepted query parameters:
- `id`: {comma separated hex strings} A list of user IDs. Will ignore all queries except ascending and orderBy if this is present.
- `name`: {comma separated strings} A list of names. Non-exact, case-insensitive match by default.
- `nameExact`: {Y/y} A flag that sets whether the database query should match the names exactly. Set this flag to Y/y only if you want to get exact matches
- `userHandle`: {comma separated strings} A list of user handles. Non-exact, case-insensitive match by default.
- `userHandleExact`: {Y/y} A flag that sets whether the query should match the userHandles exactly or not. Set this flag to Y/y only if you want to get exact matches
- `limit`: {int}: Limit on the number of results returned. Default (and max) of 100.
- `ascending`: {Y/y}: Whether to return the results in ascending order or not.
- `orderBy`: {userHandle/name}: Whether to order the name by userHandle or name. Default is userHandle. The other value is treated as the tiebreaker.
- `lt`: {string}: An *exact* string that represents that values less than this string should be returned. Corresponds to the ordering key (userHandle or name)
- `gt`: {string}: An *exact* string that represents values more than this string should be returned.  Corresponds to the ordering key (userHandle or name)

##### Returns:
- `200`: List of matching users that match the query parameters.
- `400`: If at least one of the IDs passed in is invalid.
- `500`: Internal server error.

### /images endpoints

#### [GET] /getImage

Gets image by ID, if it is visible to the user.

Accepted query parameters:
- `id`: Image ID.

Returns: `(image/*)`
- `200` OK: With the provided image.
- `400`: If id is not present, or an invalid ID is passed in.
- `404`: If image is not found, or user is not authorised to view this image. (there is no difference).
___

#### [GET] /getImageMetadata

Gets the metadata (not the actual image files) of the images in the database based on the queries passed in, in chronologically descending order.

Accepted query parameters:
- `before`: UNIX time stamp representing the latest image that can be uploaded.
- `after`: UNIX time stamp repesenting the earliest image that should be fetched.
- `limit`: integer. the limit on the number of images to fetch. Default 10 if not specified.
- `user`: comma-separated string. Gets images from particular user(s).

Returns: (application/json)
- `200`: With list of images that match search criteria.
- `500`: Internal server error.
___

#### [POST] /addImage
**Accepts**: `form/multipart`
**Returns**: `application/json`

Inserts a new image record to the database, and uploads the file to our S3 bucket.

Requires user to authenticate via the Cookie header with their JWT. Total form size has a limit of 10MB.

##### Form fields:
- `accessLevel`: Either `public` or `private`; indicates whether this image is publicly accessible or not.
- `accessListIDs`: An array of user IDs that the image is visible to. Only relevant when image is private. 
- `caption`: Image caption.
- `file`: The image file.

##### Returns:
- `200`: Image uploaded successfully. Returns the image ID in an `id` field.
- `400`: There was an error parsing the form, file, or the client did not upload an image file.
- `500`: Internal server error.

___

#### [DELETE] /deleteImage
**Accepts**: `application/json`

Deletes the given image.

JSON body parameters:
- `id`: the image ID.

Returns
- `200`: Image deleted successfully.
- `400`: Image ID was not passed in.
- `404`: Image not found or user does not have permission to delete image. (no difference)
- `500`: Internal server error

___

#### [PATCH] /editImageACL
**Accepts**: `application/json`

Adds the selected user IDs to the image's access control list (ACL)

JSON body parameters:
- `_id`: the image ID.
- `add`: an array of strings: the user IDs to add.
- `remove`: an array of strings: user IDs to remove from database.

Returns:
- `200` OK: All users were added/removed to the ACL.
- `204` No Content: ACL was not modified.
- `400`: Invalid image ID was sent, at least one invalid user ID was passed in, same user ID was present in both add and delete lists, or both add and remove lists are empty.
- `404`: At least one user in the add and remove lists does not exist in the database, or image not found.
- `500`: Internal server error

---

#### [PATCH] /likeImage
**Accepts**: `application/json`

Adds this user to the image's like list.

JSON body parameters:
- `_id`: the image ID.

Returns:
- `200`: Image liked successfully.
- `400`: id not passed in, or invalid id passed in
- `404`: Image not found, or user not authorised to view image. (no difference.)
- `409`: User already liked image. No-op.
___

#### [DELETE] /unlikeImage
**Accepts**: `application/json`

Removes this user from the image's like list.

JSON body parameters:
- `_id`: the image ID.

Returns:
- `200`: Image liked successfully.
- `400`: id not passed in, or invalid id passed in
- `404`: Image not found, or user not authorised to view image. (no difference.)
- `409`: User already liked image. No-op.
___

