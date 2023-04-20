package datatrans

type PaymentMethod string

// cf. https://api-reference.datatrans.ch/#tag/v1transactions/operation/status
// https://docs.datatrans.ch/docs/payment-methods
const (
	PaymentMethodACC = "ACC"
	PaymentMethodALP = "ALP"
	// PaymentMethodAPL is Apple Pay
	PaymentMethodAPL = "APL"
	// PaymentMethodAMX is American Express
	PaymentMethodAMX = "AMX"
	PaymentMethodAZP = "AZP"
	PaymentMethodBAC = "BAC"
	PaymentMethodBON = "BON"
	PaymentMethodCBL = "CBL"
	PaymentMethodCFY = "CFY"
	PaymentMethodCSY = "CSY"
	PaymentMethodCUP = "CUP"
	PaymentMethodDEA = "DEA"
	PaymentMethodDIN = "DIN"
	PaymentMethodDII = "DII"
	PaymentMethodDIB = "DIB"
	PaymentMethodDIS = "DIS"
	PaymentMethodDNK = "DNK"
	PaymentMethodECA = "ECA"
	PaymentMethodELV = "ELV"
	PaymentMethodEPS = "EPS"
	PaymentMethodESY = "ESY"
	PaymentMethodGFT = "GFT"
	PaymentMethodGPA = "GPA"
	PaymentMethodHPC = "HPC"
	PaymentMethodINT = "INT"
	PaymentMethodJCB = "JCB"
	PaymentMethodJEL = "JEL"
	PaymentMethodKLN = "KLN"
	PaymentMethodMAU = "MAU"
	PaymentMethodMDP = "MDP"
	PaymentMethodMFA = "MFA"
	PaymentMethodMFX = "MFX"
	PaymentMethodMPX = "MPX"
	PaymentMethodMYO = "MYO"
	PaymentMethodPAP = "PAP"
	// PaymentMethodPAY is Google Pay
	PaymentMethodPAY = "PAY"
	PaymentMethodPEF = "PEF"
	PaymentMethodPFC = "PFC"
	PaymentMethodPSC = "PSC"
	PaymentMethodREK = "REK"
	PaymentMethodSAM = "SAM"
	PaymentMethodSWB = "SWB"
	PaymentMethodSCX = "SCX"
	PaymentMethodSWP = "SWP"
	// PaymentMethodTWI is Twint
	PaymentMethodTWI = "TWI"
	PaymentMethodUAP = "UAP"
	PaymentMethodVIS = "VIS"
	PaymentMethodWEC = "WEC"
	PaymentMethodSWH = "SWH"
	PaymentMethodVPS = "VPS"
	PaymentMethodMBP = "MBP"
	PaymentMethodGEP = "GEP"
)

var (
	// AllPaymentMethods represents the list of all valid types
	AllPaymentMethods = []PaymentMethod{
		PaymentMethodACC,
		PaymentMethodALP,
		PaymentMethodAPL,
		PaymentMethodAMX,
		PaymentMethodAZP,
		PaymentMethodBAC,
		PaymentMethodBON,
		PaymentMethodCBL,
		PaymentMethodCFY,
		PaymentMethodCSY,
		PaymentMethodCUP,
		PaymentMethodDEA,
		PaymentMethodDIN,
		PaymentMethodDII,
		PaymentMethodDIB,
		PaymentMethodDIS,
		PaymentMethodDNK,
		PaymentMethodECA,
		PaymentMethodELV,
		PaymentMethodEPS,
		PaymentMethodESY,
		PaymentMethodGFT,
		PaymentMethodGPA,
		PaymentMethodHPC,
		PaymentMethodINT,
		PaymentMethodJCB,
		PaymentMethodJEL,
		PaymentMethodKLN,
		PaymentMethodMAU,
		PaymentMethodMDP,
		PaymentMethodMFA,
		PaymentMethodMFX,
		PaymentMethodMPX,
		PaymentMethodMYO,
		PaymentMethodPAP,
		PaymentMethodPAY,
		PaymentMethodPEF,
		PaymentMethodPFC,
		PaymentMethodPSC,
		PaymentMethodREK,
		PaymentMethodSAM,
		PaymentMethodSWB,
		PaymentMethodSCX,
		PaymentMethodSWP,
		PaymentMethodTWI,
		PaymentMethodUAP,
		PaymentMethodVIS,
		PaymentMethodWEC,
		PaymentMethodSWH,
		PaymentMethodVPS,
		PaymentMethodMBP,
		PaymentMethodGEP,
	}
)

// String returns the string representation
func (p PaymentMethod) String() string {
	return string(p)
}

// Valid check if the given value is included
func (p PaymentMethod) Valid() bool {
	for _, v := range AllPaymentMethods {
		if v == p {
			return true
		}
	}
	return false
}

// Is returns true if status type equals x
func (p PaymentMethod) Is(x PaymentMethod) bool {
	return x != "" && x == p
}
