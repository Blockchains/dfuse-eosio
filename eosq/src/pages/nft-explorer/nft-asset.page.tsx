import * as React from "react"
import { useSingleNFT, NFT } from "../../hooks/nft"
import styled from "@emotion/styled"

const Card = styled.table`
  max-width: 200px;
  margin: 50px;
  .imageContainer {
    width: 200px;
    height: 300px;
    display: flex;
    align-items: center;
    justify-content: center;
    img {
      max-width: 100%;
      max-height: 100%;
      width: auto;
      height: auto;
    }
  }
`

export const NftAssetPage: React.FC<{ assetId: string }> = ({ assetId }) => {
  const asset: NFT | undefined = useSingleNFT(assetId)
  if (!asset) return <div>asset not found</div>

  const { id, owner, author, category, idata, mdata } = asset
  let imageLink
  if (!mdata || !mdata.img) {
    imageLink = "/images/not-found.png"
  } else if (mdata.image.includes("http")) {
    imageLink = mdata.img
  } else {
    imageLink = `https://ipfs.io/ipfs/${mdata.img}`
  }
  return (
    <Card>
      <tbody>
        <div className="imageContainer">
          <img src={imageLink} alt={mdata.name!} />
        </div>
        <tr>ID: {id}</tr>
        <tr>Owner: {owner}</tr>
        <tr>Author: {author}</tr>
        <tr>Category: {category}</tr>
        <tr>Metadata: {JSON.stringify(mdata)}</tr>
      </tbody>
    </Card>
  )
}
