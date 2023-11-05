import React, { useState } from 'react';
import LazyLoad from 'react-lazyload';
import cosmosGray from '../assets/images/icons/cosmos_gray.png';

function ImageWithPlaceholder({ src, alt, placeholder, ...props }) {
  const [imageSrc, setImageSrc] = useState(placeholder || cosmosGray);
  const [imageRef, setImageRef] = useState();

  const onLoad = event => {
    event.target.classList.add('loaded');
  };

  const onError = () => {
    setImageSrc(cosmosGray);
  };

  // This function will be called when the actual image is loaded
  const handleImageLoad = () => {
    if (imageRef) {
      imageRef.src = src;
    }
  };
  
  return (
    <>
      <img
        ref={setImageRef}
        {...props}
        src={imageSrc}
        alt={alt}
        onLoad={onLoad}
        onError={onError}
        // style={{ opacity: imageSrc === src ? 1 : 0, transition: 'opacity 0.5s ease-in-out' }}
      />
      {/* This image will load the actual image and then handleImageLoad will be triggered */}
      <img
        {...props}
        src={src}
        alt={alt}
        style={{ display: 'none' }} // Hide this image
        onLoad={handleImageLoad}
        onError={onError}
      />
    </>
  );
}

export default ImageWithPlaceholder;