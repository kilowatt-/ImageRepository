
import React, { useState } from 'react';
import Dialog from "@material-ui/core/Dialog";
import Button from "@material-ui/core/Button";
import DialogContent from "@material-ui/core/DialogContent";
import DialogContentText from "@material-ui/core/DialogContentText";
import TextField from "@material-ui/core/TextField";
import Grid from "@material-ui/core/Grid";
import axios from "axios";
import {API_CONFIG} from "../../config/api";

const UploadModal = ({open, handleClose}) => {
    const fileInput = React.createRef();
    const [errorText, setErrorText] = useState("");

    const [file, setFile] = useState(null);
    const [caption, setCaption] = useState("");
    const [uploading, setUploading] = useState(false);

    const reset = () => {
        setFile(null);
        setUploading(false);
        setCaption("");
        setErrorText("");
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
            const response = await axios.put(`${API_CONFIG.base_url}/images/addImage`, formData, config);
        } catch (e) {
            console.log(e);
        } finally {
            setUploading(false);
        }
    }

    const handleSelectImage = (e) => {
        const file = fileInput.current.files[0];
        if (file) {
            if (file.size <= 9000000) {
                setErrorText("");
                setFile(fileInput.current.files[0]);
            } else {
                setFile(null);
                setErrorText("Selected image exceeds size limit (9MB)")
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
                                accept="image/*"
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
                            <TextField fullWidth
                                       multiline
                                       id="caption"
                                       label="caption"
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
                            >Submit</Button>
                        </Grid>
                    </>) : null
                    }
                        <p>{errorText}</p>

                </form>
            </DialogContent>
            </Grid>
        </Dialog>
    )
}

export default UploadModal;