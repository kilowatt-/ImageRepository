
import React, { useState } from 'react';
import Dialog from "@material-ui/core/Dialog";
import Button from "@material-ui/core/Button";
import DialogContent from "@material-ui/core/DialogContent";
import DialogContentText from "@material-ui/core/DialogContentText";
import TextField from "@material-ui/core/TextField";
import Grid from "@material-ui/core/Grid";
import axios from "axios";
import {API_CONFIG} from "../../config/api";
import CircularProgress from "@material-ui/core/CircularProgress";

const UploadModal = ({open, handleClose}) => {
    const fileInput = React.createRef();
    const [errorMessage, setErrorMessage] = useState("");

    const [file, setFile] = useState(null);
    const [caption, setCaption] = useState("");
    const [uploading, setUploading] = useState(false);

    const fileURL = (file) ? URL.createObjectURL(file) : null;

    const reset = () => {
        setFile(null);
        setUploading(false);
        setCaption("");
        setErrorMessage("");
    }

    const handleSubmit = async (e) => {
        e.preventDefault();
        const formData = new FormData();
        formData.append('file', file);
        formData.append('caption', caption);

        const config = {
            headers: {
                'content-type': 'multipart/form-data',
            }
        }

        setUploading(true);

        try {
            axios.defaults.withCredentials = true;
            await axios.put(`${API_CONFIG.base_url}/images/addImage`, formData, config);
            handleClose(true);
            reset();
        } catch (err) {
            if (err.response && err.response.data) {
                setErrorMessage(err.response.data);
            } else {
                setErrorMessage("Unknown error occurred while uploading file")
            }
        } finally {
            setUploading(false);
        }
    }

    const handleSelectImage = (e) => {
        const file = fileInput.current.files[0];
        if (file) {
            if (file.size <= 9000000) {
                setErrorMessage("");
                setFile(fileInput.current.files[0]);
            } else {
                setFile(null);
                setErrorMessage("Selected image exceeds size limit (9MB)")
            }
        }
    }

    return (
        <Dialog open={open} onClose={() => {reset(); handleClose()}} fullWidth={true}>
            <Grid container spacing={3}>
            <DialogContent>
                <DialogContentText>Upload image</DialogContentText>
                <form onSubmit={handleSubmit} noValidate>

                    {!file ?
                        (
                        <Grid item xs={12}>
                            <input
                                accept="image/gif, image/jpeg, image/bmp, image/gif, image/webp"
                                style={{ display: 'none' }}
                                id="image-upload"
                                type="file"
                                ref={fileInput}
                                onChange={handleSelectImage}
                            />
                            <label htmlFor="image-upload">
                                <Button variant="contained" component="span">
                                    Select image (maximum 9MB)
                                </Button>
                            </label>
                        </Grid>
                        ) :(
                            <Grid item xs={12}>
                            <Button variant="contained" component="span" onClick={() => setFile(null)}>
                                    Delete
                                </Button>
                        </Grid>
                        )}
                    {file ?(
                    <>
                        <Grid item xs={12}>
                            <img src={fileURL} />
                        </Grid>
                        <Grid item xs={12}>
                            <TextField fullWidth
                                       multiline
                                       id="caption"
                                       label="Caption"
                                       variant="outlined"
                                       autoFocus
                                       value={caption}
                                       onChange={(e) => setCaption(e.target.value)}/>
                        </Grid>
                        <Grid item xs={12}>
                            <Button
                                type="submit"
                                fullWidth
                                variant="contained"
                                color="primary"
                                disabled={uploading}
                            >{uploading ? <CircularProgress color={"secondary"}/> : "Upload"}</Button>
                        </Grid>
                    </>) : null
                    }
                        <p>{errorMessage}</p>

                </form>
            </DialogContent>
            </Grid>
        </Dialog>
    )
}

export default UploadModal;