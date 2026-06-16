package memory

import (
	"GarageSaleAPI/domain/sale"
	"reflect"
	"testing"
)

func TestInMemorySaleRepository_AddSale(t *testing.T) {
	type fields struct {
		SaleList []sale.Sale
	}
	type args struct {
		sale sale.Sale
	}

	validSale := sale.CreateSale(
		"123e4567-e89b-12d3-a456-426614174000",
		"Sale",
		"",
	)

	tests := []struct {
		name        string
		fields      fields
		args        args
		wantErr     bool
		wantErrText string
	}{
		{
			name: "add sale",
			fields: fields{
				[]sale.Sale{},
			},
			args: args{
				sale: validSale,
			},
			wantErr: false,
		},
		{
			name: "add duplicate sale",
			fields: fields{
				[]sale.Sale{validSale},
			},
			args: args{
				sale: validSale,
			},
			wantErr:     true,
			wantErrText: "sale already exists",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &InMemorySaleRepository{
				saleList: tt.fields.SaleList,
			}

			err := repo.AddSale(tt.args.sale)

			if (err != nil) != tt.wantErr ||
				((err != nil) && err.Error() != tt.wantErrText) {
				t.Errorf("AddSale() error = %v, wantErr %v\ntext = %v, textErr = %v",
					err, tt.wantErr, err.Error(), tt.wantErrText)
			}
		})
	}
}

func TestInMemorySaleRepository_GetSaleById(t *testing.T) {
	type fields struct {
		SaleList []sale.Sale
	}
	type args struct {
		id string
	}

	validId := "123e4567-e89b-12d3-a456-426614174000"
	validSale := sale.CreateSale(
		validId,
		"Sale",
		"",
	)

	tests := []struct {
		name        string
		fields      fields
		args        args
		want        *sale.Sale
		wantErr     bool
		wantErrText string
	}{
		{
			name: "get existing sale",
			fields: fields{
				[]sale.Sale{validSale},
			},
			args: args{
				validId,
			},
			want:    &validSale,
			wantErr: false,
		},
		{
			name: "get nonexistent sale",
			fields: fields{
				[]sale.Sale{
					sale.CreateSale(
						"123e4567-e89b-12d3-a456-426614485967",
						"diff sale",
						"1234 create st",
					),
				},
			},
			args: args{
				validId,
			},
			want:        nil,
			wantErr:     true,
			wantErrText: "sale not found",
		},
		{
			name: "get sale with empty list",
			fields: fields{
				[]sale.Sale{},
			},
			args: args{
				validId,
			},
			want:        nil,
			wantErr:     true,
			wantErrText: "sale not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &InMemorySaleRepository{
				saleList: tt.fields.SaleList,
			}

			got, err := repo.GetSaleById(tt.args.id)

			if (err != nil) != tt.wantErr ||
				((err != nil) && err.Error() != tt.wantErrText) {
				t.Errorf("GetSaleById() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSaleById() got = %v, want %v", got, tt.want)
			}
		})
	}
}
