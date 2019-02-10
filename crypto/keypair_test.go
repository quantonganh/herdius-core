package crypto

// import (
// 	"testing"

// 	"github.com/golang/mock/gomock"
// 	"github.com/herdius/herdius-core/p2p/crypto/mocks"
// )

// var (
// 	privateKey    []byte
// 	privateKeyHex string
// 	publicKey     []byte
// 	publicKeyHex  string
// 	message       []byte
// 	hashed        []byte
// 	signature     []byte
// )

// func init() {
// 	// mock inputs
// 	privateKey = []byte("1234567890")
// 	privateKeyHex = "31323334353637383930"
// 	publicKey = []byte("12345678901234567890")
// 	publicKeyHex = "3132333435363738393031323334353637383930"

// 	message = []byte("test message")
// 	hashed = []byte("hashed test message")
// 	signature = []byte("signed test message")
// }

// func TestKeyPair(t *testing.T) {
// 	t.Parallel()

// 	mockCtrl := gomock.NewController(t)
// 	defer mockCtrl.Finish()

// 	sp := mocks.NewMockSignaturePolicy(mockCtrl)
// 	hp := mocks.NewMockHashPolicy(mockCtrl)

// 	// setup expected mock return values
// 	sp.EXPECT().PrivateKeySize().Return(len(privateKey)).AnyTimes()
// 	sp.EXPECT().PublicKeySize().Return(len(publicKey)).AnyTimes()
// 	sp.EXPECT().Sign(privateKey, hashed).Return(signature).AnyTimes()
// 	sp.EXPECT().Verify(publicKey, hashed, signature).Return(true).Times(1)

// 	hp.EXPECT().HashBytes(message).Return(hashed).AnyTimes()
// }
