package main

import (
	"flag"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"os"
	"os/signal"

	"github.com/3d0c/gmf"
)

var isKilled = false

func printVideo(srcPath string) {
	var swsctx *gmf.SwsCtx

	inputCtx, err := gmf.NewInputCtx(srcPath)
	if err != nil {
		log.Panicf("error: Error creating context - %s\n", err)
	}
	defer inputCtx.Free()

	srcVideoStream, err := inputCtx.GetBestStream(gmf.AVMEDIA_TYPE_VIDEO)
	if err != nil {
		log.Printf("error: No video stream found in '%s'\n", srcPath)
		return
	}

	codec, err := gmf.FindEncoder(gmf.AV_CODEC_ID_RAWVIDEO)
	if err != nil {
		log.Panicf("error: %s\n", err)
	}

	cc := gmf.NewCodecCtx(codec)
	defer gmf.Release(cc)

	cc.SetTimeBase(gmf.AVR{Num: 1, Den: 1})
	cc.SetPixFmt(gmf.AV_PIX_FMT_RGBA).SetWidth(srcVideoStream.CodecCtx().Width()).SetHeight(srcVideoStream.CodecCtx().Height())
	if codec.IsExperimental() {
		cc.SetStrictCompliance(gmf.FF_COMPLIANCE_EXPERIMENTAL)
	}

	if err := cc.Open(nil); err != nil {
		log.Panic("error: ", err)
	}
	defer cc.Free()

	ist, err := inputCtx.GetStream(srcVideoStream.Index())
	if err != nil {
		log.Panicf("error: Error getting stream - %s\n", err)
	}
	defer ist.Free()

	icc := srcVideoStream.CodecCtx()
	if swsctx, err = gmf.NewSwsCtx(icc.Width(), icc.Height(), icc.PixFmt(), cc.Width(), cc.Height(), cc.PixFmt(), gmf.SWS_BICUBIC); err != nil {
		log.Panic("error: ", err)
	}
	defer swsctx.Free()

	var (
		pkt        *gmf.Packet
		frames     []*gmf.Frame
		drain      int = -1
		frameCount int = 0
	)

	for {
		if drain >= 0 {
			break
		}

		pkt, err = inputCtx.GetNextPacket()
		if err != nil && err != io.EOF {
			if pkt != nil {
				pkt.Free()
			}
			log.Printf("error: error getting next packet - %s", err)
			break
		} else if err != nil && pkt == nil {
			drain = 0
		}

		if pkt != nil && pkt.StreamIndex() != srcVideoStream.Index() {
			continue
		}

		frames, err = ist.CodecCtx().Decode(pkt)
		if err != nil {
			log.Printf("error: Fatal error during decoding - %s\n", err)
			break
		}

		if len(frames) == 0 && drain < 0 {
			continue
		}

		if frames, err = gmf.DefaultRescaler(swsctx, frames); err != nil {
			log.Panic("error: ", err)
		}

		encode(cc, frames, drain)

		for i := range frames {
			frames[i].Free()
			frameCount++
		}

		if pkt != nil {
			pkt.Free()
			pkt = nil
		}

		if isKilled {
			break
		}
	}

	for i := 0; i < inputCtx.StreamsCnt(); i++ {
		st, _ := inputCtx.GetStream(i)
		st.CodecCtx().Free()
		st.Free()
	}
}

func encode(cc *gmf.CodecCtx, frames []*gmf.Frame, drain int) {
	packets, err := cc.Encode(frames, drain)
	if err != nil {
		log.Panicf("error: Error encoding - %s\n", err)
	}
	if len(packets) == 0 {
		return
	}

	for _, p := range packets {
		width, height := cc.Width(), cc.Height()

		img := new(image.RGBA)
		img.Pix = p.Data()
		img.Stride = 4 * width
		img.Rect = image.Rect(0, 0, width, height)

		resizedImg := ResizeRGBA(img)
		fmt.Print("\x1b[1;1H")
		PrintImg(resizedImg)

		p.Free()
		if isKilled {
			break
		}
	}
}

func main() {
	flag.Parse()
	if flag.NArg() != 1 {
		log.Println("error: no src")
		return
	}

	src, err := os.Open(flag.Arg(0))
	if err != nil {
		log.Println("error:", err)
		return
	}
	defer src.Close()

	img, _, err := image.Decode(src)
	if err == nil {
		resizedImg := ResizeImg(&img)
		PrintImg(resizedImg)
		return
	}

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	go func() {
		<-quit
		isKilled = true
		close(quit)
	}()

	fmt.Print("\x1b[?25l")
	printVideo(flag.Arg(0))
	fmt.Print("\x1b[?25h\x1b[0m")
	return
}
